package challenge

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum/go-ethereum/common"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup/status"

	"github.com/ethereum-optimism/optimism/op-dachallenger/config"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/event"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/log"
)

type CoordinatorMetricer interface {
	RecordActedL1Block(n uint64)
}

type Coordinator struct {
	challenger *scheduler.ChallengeScheduler
	resolver   *scheduler.ResolveScheduler
	unlocker   *scheduler.BondUnlockScheduler
	withdrawer *scheduler.WithdrawScheduler

	plasmaFetcher derive.PlasmaInputFetcher
	l1Fetcher     plasma.L1Fetcher
	factory       *derive.DataSourceFactory

	comm plasma.CommitmentData
	id   eth.L1BlockRef
	batcherAddr common.Address

	l1FinalizedSig chan eth.L1BlockRef
	l1NewHeadSig   chan eth.L1BlockRef
	quit           chan struct{}

	logger log.Logger
	m      CoordinatorMetricer

	// lastScheduledBlockNum is the highest block number that the coordinator has seen and scheduled jobs.
	lastScheduledBlockNum uint64
}

func NewCoordinator(logger log.Logger, m CoordinatorMetricer, c *scheduler.ChallengeScheduler, r *scheduler.ResolveScheduler,
	u *scheduler.BondUnlockScheduler, w *scheduler.WithdrawScheduler, damgr derive.PlasmaInputFetcher,
	l1Fetcher plasma.L1Fetcher, factory *derive.DataSourceFactory, batcherAddr common.Address) *Coordinator {
	return &Coordinator{
		logger:         logger,
		m:              m,
		challenger:     c,
		resolver:       r,
		unlocker:       u,
		withdrawer:     w,
		plasmaFetcher:  damgr,
		factory: 		factory,
		l1Fetcher:      l1Fetcher,
		l1FinalizedSig: make(chan eth.L1BlockRef),
		quit:           make(chan struct{}),
		batcherAddr: 	batcherAddr,
	}
}

func (c *Coordinator) OnNewL1Finalized(ctx context.Context, finalized eth.L1BlockRef) {
	select {
	case <-ctx.Done():
		return
	case c.l1FinalizedSig <- finalized:
		return
	}
}

func (c *Coordinator) OnNewL1Head(ctx context.Context, head eth.L1BlockRef) {
	select {
	case <-ctx.Done():
		return
	case c.l1NewHeadSig <- head:
		return
	}
}

// Questions:
// Is it sufficient to only attempt to challenge the commitment once? E.g. on every new finalized block
// we identify new commitments in that block, and check if we can retrieve the input. IFF we can't then we
// create a new challenge. If we can, we do not and do not come back to check at a later time.

// Similarly, if we see a new challenge in a new finalized block do we only attempt to resolve once or do we retry
// on every subsequent block until we can either resolve or the resolveWindow expires?
func (c *Coordinator) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.quit:
			return
		case <-c.l1FinalizedSig:
			// Is it sufficient to only watch the fina
			// Process origin syncs the challenge contract events and updates the local challenge states
			// before we can proceed to fetch the input data. This function can be called multiple times
			// for the same origin and noop if the origin was already processed. It is also called if
			// there is not commitment in the current origin.
			if err := c.plasmaFetcher.AdvanceL1Origin(ctx, c.l1Fetcher, c.id.ID()); err != nil {
				if errors.Is(err, plasma.ErrReorgRequired) {
					return nil, derive.NewResetError(fmt.Errorf("new expired challenge"))
				}
				return nil, derive.NewTemporaryError(fmt.Errorf("failed to advance plasma L1 origin: %w", err))
			}

			if c.comm == nil {
				// the l1 source returns the input commitment for the batch.
				data, err := c.src.Next(ctx)
				if err != nil {
					return nil, err
				}

				if len(data) == 0 {
					return nil, derive.NotEnoughData
				}
				// If the tx data type is not plasma, we forward it downstream to let the next
				// steps validate and potentially parse it as L1 DA inputs.
				if data[0] != plasma.TxDataVersion1 {
					return data, nil
				}

				// validate batcher inbox data is a commitment.
				// strip the transaction data version byte from the data before decoding.
				comm, err := plasma.DecodeCommitmentData(data[1:])
				if err != nil {
					c.logger.Warn("invalid commitment", "commitment", data, "err", err)
					return nil, derive.NotEnoughData
				}
				c.comm = comm
			}
			// look at the L1BlockRef to figure out what to schedule
			// use the commitment to fetch the input from the plasma DA provider.
			data, err := c.plasmaFetcher.GetInput(ctx, c.l1Fetcher, c.comm, c.id)
		case <-c.l1NewHeadSig:
			// everytime we receive a new head, we need to retry our challenge and response logic
			// retry challenge: check our commitments

		}
	}
}

// TODO: we also need to handle retry logic for the entire challenge_window and resolve_window

func (c *Coordinator) Close() error {
	c.quit <- struct{}{}
	return nil
}
