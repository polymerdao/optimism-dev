package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/log"
)

type ChallengeContract interface {
	Challenge(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error)
}

type Challenger struct {
	logger   log.Logger
	contract ChallengeContract
	txSender TxSender
}

var _ Challenging = (*Challenger)(nil)

func NewChallenger(l log.Logger, contract ChallengeContract, txSender TxSender) *Challenger {
	return &Challenger{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (c *Challenger) ChallengeCommitments(ctx context.Context, commitments []types.CommitmentArg) (err error) {
	for _, comm := range commitments {
		err = errors.Join(err, c.challengeCommitment(ctx, comm))
	}
	return err
}

func (c *Challenger) challengeCommitment(ctx context.Context, commitment types.CommitmentArg) error {
	c.logger.Debug("Attempting to challenge commitment for", "blockheight", commitment.ChallengedBlockNumber,
		"commitment", commitment.ChallengedCommitment)

	candidate, err := c.contract.Challenge(ctx, commitment)
	if err != nil {
		return fmt.Errorf("failed to create candidate challenge tx: %w", err)
	}

	if err = c.txSender.SendAndWaitSimple("challenge da commitment", candidate); err != nil {
		return fmt.Errorf("failed to challenge da commitment: %w", err)
	}

	return nil
}
