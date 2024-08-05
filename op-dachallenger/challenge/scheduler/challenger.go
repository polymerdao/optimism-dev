package scheduler

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type ChallengeContractCreator func(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (types.ChallengeContract, error)

type Challenger struct {
	logger   log.Logger
	contract types.ChallengeContract
	txSender types.TxSender
}

var _ Challenging = (*Challenger)(nil)

func NewChallenger(l log.Logger, contract types.ChallengeContract, txSender types.TxSender) *Challenger {
	return &Challenger{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (c *Challenger) ChallengeCommitment(ctx context.Context, commitment types.CommitmentArg) (err error) {
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
