package scheduler

import (
	"context"
	"sync"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum/go-ethereum/log"
)

type Resolving interface {
	ResolveChallenges(ctx context.Context, resolveData []types.ResolveData) error
}

type ResolveScheduler struct {
	log      log.Logger
	metrics  ResolveSchedulerMetrics
	ch       chan resolveMessage
	resolver Resolving
	cancel   func()
	wg       sync.WaitGroup
}

type ResolveSchedulerMetrics interface {
	RecordDAChallenge()
	RecordDAChallengeFailed() // TODO: HERE
}

type resolveMessage struct {
	blockNumber uint64
	resolveData []types.ResolveData
}

func NewResolveScheduler(logger log.Logger, metrics ResolveSchedulerMetrics, resolver Resolving) *ResolveScheduler {
	return &ResolveScheduler{
		log:      logger,
		metrics:  metrics,
		ch:       make(chan resolveMessage, 1),
		resolver: resolver,
	}
}

func (s *ResolveScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
}

func (s *ResolveScheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *ResolveScheduler) run(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			if err := s.resolver.ResolveChallenges(ctx, msg.resolveData); err != nil {
				s.metrics.RecordDAChallengeFailed()
				s.log.Error("Failed to resolve challenges", "blockNumber", msg.blockNumber, "err", err)
			} else {
				s.metrics.RecordDAChallenge()
			}
		}
	}
}

func (s *ResolveScheduler) Schedule(blockNumber uint64, resolveData []types.ResolveData) error {
	select {
	case s.ch <- resolveMessage{blockNumber, resolveData}:
	}
	return nil
}
