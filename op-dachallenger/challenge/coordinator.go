package challenge

import (
	"context"
	"errors"
	"fmt"

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

	damgr *plasma.DA

	l1FinalizedSig chan eth.L1BlockRef
	quit           chan struct{}

	logger log.Logger
	m      CoordinatorMetricer

	// lastScheduledBlockNum is the highest block number that the coordinator has seen and scheduled jobs.
	lastScheduledBlockNum uint64
}

func NewCoordinator(logger log.Logger, m CoordinatorMetricer, c *scheduler.ChallengeScheduler, r *scheduler.ResolveScheduler,
	u *scheduler.BondUnlockScheduler, w *scheduler.WithdrawScheduler, damgr *plasma.DA) *Coordinator {
	return &Coordinator{
		logger:         logger,
		m:              m,
		challenger:     c,
		resolver:       r,
		unlocker:       u,
		withdrawer:     w,
		damgr:          damgr,
		l1FinalizedSig: make(chan eth.L1BlockRef),
		quit:           make(chan struct{}),
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

func (c *Coordinator) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.quit:
			return
		case <-c.l1FinalizedSig:
			// Process origin syncs the challenge contract events and updates the local challenge states
			// before we can proceed to fetch the input data. This function can be called multiple times
			// for the same origin and noop if the origin was already processed. It is also called if
			// there is not commitment in the current origin.
			if err := c.fetcher.AdvanceL1Origin(ctx, c.l1, c.id.ID()); err != nil {
				if errors.Is(err, plasma.ErrReorgRequired) {
					return nil, NewResetError(fmt.Errorf("new expired challenge"))
				}
				return nil, NewTemporaryError(fmt.Errorf("failed to advance plasma L1 origin: %w", err))
			}

			if c.comm == nil {
				// the l1 source returns the input commitment for the batch.
				data, err := c.src.Next(ctx)
				if err != nil {
					return nil, err
				}

				if len(data) == 0 {
					return nil, NotEnoughData
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
					return nil, NotEnoughData
				}
				c.comm = comm
			}
			// look at the L1BlockRef to figure out what to schedule
			// use the commitment to fetch the input from the plasma DA provider.
			data, err := c.damgr.GetInput(ctx, c.l1, c.comm, c.id)
		}
	}
}

func (c *Coordinator) Close() error {
	c.quit <- struct{}{}
	return nil
}
