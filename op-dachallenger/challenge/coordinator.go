package scheduler

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/log"
)

type CoordinatorMetricer interface {
	RecordActedL1Block(n uint64)
}

type Coordinator struct {
	challenger *ChallengeScheduler
	resolver   *ResolveScheduler
	unlocker   *BondUnlockScheduler
	withdrawer *WithdrawScheduler

	l1FinalizedSig chan eth.L1BlockRef
	quit           chan struct{}

	logger log.Logger
	m      CoordinatorMetricer

	// lastScheduledBlockNum is the highest block number that the coordinator has seen and scheduled jobs.
	lastScheduledBlockNum uint64
}

func NewCoordinator(logger log.Logger, m CoordinatorMetricer, c *ChallengeScheduler, r *ResolveScheduler,
	u *BondUnlockScheduler, w *WithdrawScheduler) *Coordinator {
	return &Coordinator{
		logger:         logger,
		m:              m,
		challenger:     c,
		resolver:       r,
		unlocker:       u,
		withdrawer:     w,
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
			// look at the L1BlockRef to figure out what to schedule
		}
	}
}

func (c *Coordinator) Close() error {
	c.quit <- struct{}{}
	return nil
}
