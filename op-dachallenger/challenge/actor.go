package challenge

import (
	"context"
	"errors"
	"io"
	"math/big"
	"sync"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/log"
)

type ActorMetricer interface {
	RecordLastActedL1Block(n uint64)
	RecordActedL1Ref(ref eth.L1BlockRef)
}

type Actor struct {
	resolver   *scheduler.ResolveScheduler
	challenger *scheduler.ChallengeScheduler

	plasmaFetcher types.PlasmaFetcher
	l1Fetcher     derive.L1Fetcher
	blobsFetcher  *sources.L1BeaconClient
	rollupCfg     *rollup.Config
	dsConfig      derive.DataSourceConfig

	l1FinalizedSig chan eth.L1BlockRef
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	logger  log.Logger
	metrics ActorMetricer
}

func NewActor(logger log.Logger, m ActorMetricer, r *scheduler.ResolveScheduler, c *scheduler.ChallengeScheduler,
	damgr types.PlasmaFetcher, l1Fetcher derive.L1Fetcher, blobsFetcher *sources.L1BeaconClient,
	rollCfg *rollup.Config) (*Actor, error) {
	if r == nil && c == nil {
		return nil, errors.New("actor needs to be configured with a challenger or a resolver, or both")
	}
	dsConfig := derive.DataSourceConfig{
		L1Signer:          rollCfg.L1Signer(),
		BatchInboxAddress: rollCfg.BatchInboxAddress,
		PlasmaEnabled:     true,
	}
	return &Actor{
		logger:         logger,
		metrics:        m,
		resolver:       r,
		challenger:     c,
		plasmaFetcher:  damgr,
		blobsFetcher:   blobsFetcher,
		rollupCfg:      rollCfg,
		dsConfig:       dsConfig,
		l1Fetcher:      l1Fetcher,
		l1FinalizedSig: make(chan eth.L1BlockRef),
	}, nil
}

func (a *Actor) OnNewL1Finalized(ctx context.Context, finalized eth.L1BlockRef) {
	select {
	case <-ctx.Done():
		return
	case a.l1FinalizedSig <- finalized:
		return
	}
}

// Questions:
// Is it sufficient to only attempt to challenge the commitment once? E.g. on every new finalized block
// we identify new commitments in that block, and check if we can retrieve the input. IFF we can't then we
// create a new challenge. If we can, we do not and do not come back to check at a later time.

// Similarly, if we see a new challenge in a new finalized block do we only attempt to resolve once or do we retry
// on every subsequent block until we can either resolve or the resolveWindow expires?

// Start the DA challenge/resolve loop
func (a *Actor) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.resolver.Start(ctx)
	a.wg.Add(1)
	go a.run(ctx)
}

func (a *Actor) run(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ref := <-a.l1FinalizedSig:
			// TODO: consider also handling retry logic for the entire challenge_window
			// TODO: ensure it is OK to use latest finalized as L1 origin
			// it should be fine as long as resolve_window and challenge_window are significantly longer than 64 slots
			var src derive.DataIter
			if a.rollupCfg.EcotoneTime != nil && ref.Time >= *a.rollupCfg.EcotoneTime {
				if a.blobsFetcher == nil {
					a.logger.Error("ecotone upgrade active but beacon endpoint not configured")
					return
				}
				src = derive.NewBlobDataSource(ctx, a.logger, a.dsConfig, a.l1Fetcher, a.blobsFetcher, ref, a.rollupCfg.BatchInboxAddress)
			} else {
				src = derive.NewCalldataSource(ctx, a.logger, a.dsConfig, a.l1Fetcher, ref, a.rollupCfg.BatchInboxAddress)
			}
			src = derive.NewPlasmaDataSource(a.logger, src, a.l1Fetcher, a.plasmaFetcher, ref)

			if err := a.plasmaFetcher.AdvanceL1Origin(ctx, a.l1Fetcher, ref.ID()); err != nil {
				if !errors.Is(err, plasma.ErrReorgRequired) {
					a.logger.Error("failed to advance plasma L1 origin: ", err)
					return
				}
			}
			for {
				// the l1 source returns the input commitment for the batch.
				data, err := src.Next(ctx)
				if err == io.EOF {
					// no (more) commitments to process for this block ref, break the loop
					break
				} else if err != nil {
					// critical failure upstream
					a.logger.Error("failed to return input commitment for batch: ", err)
					// TODO return this error
					return
				}

				if len(data) == 0 {
					// critical failure upstream if we were returned no data without an io.EOF error
					a.logger.Error(derive.NotEnoughData.Error())
					// TODO: return this error
					return
				}

				// If the tx data type is not plasma, we skip it
				if data[0] != plasma.TxDataVersion1 {
					continue
				}

				// validate batcher inbox data is a commitment.
				// strip the transaction data version byte from the data before decoding.
				comm, err := plasma.DecodeCommitmentData(data[1:])
				if err != nil {
					// invalid commitment, don't need to do anything with it
					a.logger.Warn("invalid commitment", "commitment", data, "err", err)
					continue
				}

				// use the commitment to fetch the input from the plasma DA provider.
				daData, err := a.plasmaFetcher.GetInput(ctx, a.l1Fetcher, comm, ref)
				switch {
				case errors.Is(err, plasma.ErrActiveChallenge):
					// we were unable to fetch the data but there is already an active challenge, do nothing
					// separating this from the default because in the future we may want to handle this case by
					// periodically trying to fetch the missing data and resolve the challenge
				case errors.Is(err, plasma.ErrPendingChallenge):
					// TODO: consider also handling retry logic for the entire challenge_window
					// if we were unable to fetch the data but there is no active challenge, then we should create one
					if a.challenger != nil {
						a.logger.Info("challenging altDA commitment with missing data",
							"commitment height", ref.Number,
							"commitment", comm.Encode())
						if err := a.challenger.Schedule(ref.Number, types.CommitmentArg{
							ChallengedBlockNumber: new(big.Int).SetUint64(ref.Number),
							ChallengedCommitment:  comm.Encode(),
						}); err != nil {
							// TODO: better error handling e.g. we need to retry
							a.logger.Error(err.Error())
						}
						a.metrics.RecordLastActedL1Block(ref.Number)
						a.metrics.RecordActedL1Ref(ref)
					}
				case errors.Is(err, nil):
					// if we were able to fetch the data but there is an active challenge, then we should resolve it
					status := a.plasmaFetcher.GetChallengeStatus(comm, ref.Number)
					if status == plasma.ChallengeActive {
						a.logger.Info("resolving altDA commitment with retrieved data",
							"commitment height", ref.Number,
							"commitment", comm.Encode())
						if err := a.resolver.Schedule(ref.Number, types.ResolveData{
							CommitmentArg: types.CommitmentArg{
								ChallengedBlockNumber: new(big.Int).SetUint64(ref.Number),
								ChallengedCommitment:  comm.Encode(),
							},
							Blob: daData,
						}); err != nil {
							// TODO: better error handling e.g. we need to retry
							a.logger.Error(err.Error())
						}
						a.metrics.RecordLastActedL1Block(ref.Number)
						a.metrics.RecordActedL1Ref(ref)
					}
				default:
					// other cases where there is nothing to do: ErrPendingChallenge, ErrExpiredChallenge, ErrMissingPastWindow.
				}
			}
		}
	}
}

func (a *Actor) Close() error {
	a.cancel()
	a.wg.Wait()
	return nil
}
