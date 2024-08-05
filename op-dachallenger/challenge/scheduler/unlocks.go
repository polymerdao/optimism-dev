package scheduler

import (
	"context"
	"sync"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum/go-ethereum/log"
)

type BondUnlocker interface {
	UnlockBond(ctx context.Context, challenge types.CommitmentArg) error
}

type BondUnlockScheduler struct {
	log      log.Logger
	metrics  BondUnlockSchedulerMetrics
	ch       chan unlockMessage
	unlocker BondUnlocker
	cancel   func()
	wg       sync.WaitGroup
}

type BondUnlockSchedulerMetrics interface {
	RecordBondUnlockFailed()
	RecordBondUnlock()
}

type unlockMessage struct {
	blockNumber uint64
	challenge   types.CommitmentArg
}

func NewBondUnlockScheduler(logger log.Logger, metrics BondUnlockSchedulerMetrics, unlocker BondUnlocker) *BondUnlockScheduler {
	return &BondUnlockScheduler{
		log:      logger,
		metrics:  metrics,
		ch:       make(chan unlockMessage, 1),
		unlocker: unlocker,
	}
}

func (s *BondUnlockScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
}

func (s *BondUnlockScheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *BondUnlockScheduler) run(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			if err := s.unlocker.UnlockBond(ctx, msg.challenge); err != nil {
				s.metrics.RecordBondUnlockFailed()
				s.log.Error("Failed to claim bonds", "blockNumber", msg.blockNumber, "err", err)
			}
		}
	}
}

func (s *BondUnlockScheduler) Schedule(blockNumber uint64, challenge types.CommitmentArg) error {
	select {
	case s.ch <- unlockMessage{blockNumber, challenge}:
	default:
		s.log.Trace("Skipping game bond claim while claiming in progress")
	}
	return nil
}
