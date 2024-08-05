package scheduler

import (
	"context"
	"sync"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum/go-ethereum/log"
)

type Challenging interface {
	ChallengeCommitment(ctx context.Context, commitment types.CommitmentArg) error
}

type ChallengeScheduler struct {
	log        log.Logger
	metrics    ChallengerSchedulerMetrics
	ch         chan challengeMessage
	challenger Challenging
	cancel     func()
	wg         sync.WaitGroup
}

type ChallengerSchedulerMetrics interface {
	RecordDAChallengeFailed()
	RecordDAChallenge()
}

type challengeMessage struct {
	blockNumber uint64
	commitment  types.CommitmentArg
}

func NewChallengeScheduler(logger log.Logger, metrics ChallengerSchedulerMetrics, challenger Challenging) *ChallengeScheduler {
	return &ChallengeScheduler{
		log:        logger,
		metrics:    metrics,
		ch:         make(chan challengeMessage, 1),
		challenger: challenger,
	}
}

func (s *ChallengeScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
}

func (s *ChallengeScheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *ChallengeScheduler) run(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			if err := s.challenger.ChallengeCommitment(ctx, msg.commitment); err != nil {
				s.metrics.RecordDAChallengeFailed()
				s.log.Error("Failed to challenge commitments", "blockNumber", msg.blockNumber, "err", err)
			} else {
				s.metrics.RecordDAChallenge()
			}
		}
	}
}

func (s *ChallengeScheduler) Schedule(blockNumber uint64, commitment types.CommitmentArg) error {
	select {
	case s.ch <- challengeMessage{blockNumber, commitment}:
	}
	return nil
}
