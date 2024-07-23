package scheduler

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/log"
)

type BondWithdrawer interface {
	WithdrawBonds(ctx context.Context, height uint64) (err error)
}

type WithdrawScheduler struct {
	log        log.Logger
	metrics    WithdrawSchedulerMetrics
	ch         chan withdrawMessage
	withdrawer BondWithdrawer
	cancel     func()
	wg         sync.WaitGroup
}

type WithdrawSchedulerMetrics interface {
	RecordWithdrawFailed()
}

type withdrawMessage struct {
	blockNumber uint64
}

func NewWithdrawScheduler(logger log.Logger, metrics WithdrawSchedulerMetrics, withdrawer BondWithdrawer) *WithdrawScheduler {
	return &WithdrawScheduler{
		log:        logger,
		metrics:    metrics,
		ch:         make(chan withdrawMessage, 1),
		withdrawer: withdrawer,
	}
}

func (s *WithdrawScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
}

func (s *WithdrawScheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *WithdrawScheduler) run(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			if err := s.withdrawer.WithdrawBonds(ctx, msg.blockNumber); err != nil {
				s.metrics.RecordWithdrawFailed()
				s.log.Error("Failed to claim bonds", "blockNumber", msg.blockNumber, "err", err)
			}
		}
	}
}

func (s *WithdrawScheduler) Schedule(blockNumber uint64) error {
	select {
	case s.ch <- withdrawMessage{blockNumber}:
	default:
		s.log.Trace("Skipping game bond claim while claiming in progress")
	}
	return nil
}
