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

type ResolveContractCreator func(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (types.ResolveContract, error)

type Resolver struct {
	logger   log.Logger
	contract types.ResolveContract
	txSender types.TxSender
}

var _ Resolving = (*Resolver)(nil)

func NewResolver(l log.Logger, contract types.ResolveContract, txSender types.TxSender) *Resolver {
	return &Resolver{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (c *Resolver) ResolveChallenge(ctx context.Context, resolveData types.ResolveData) (err error) {
	c.logger.Debug("Attempting to resolve challenge for", "blockheight", resolveData.ChallengedBlockNumber,
		"commitment", resolveData.ChallengedCommitment)

	candidate, err := c.contract.Resolve(ctx, resolveData.CommitmentArg, resolveData.Blob)
	if err != nil {
		return fmt.Errorf("failed to create candidate resolve tx: %w", err)
	}

	if err = c.txSender.SendAndWaitSimple("resolve da challenge", candidate); err != nil {
		return fmt.Errorf("failed to resolve da challenge: %w", err)
	}

	return nil
}
