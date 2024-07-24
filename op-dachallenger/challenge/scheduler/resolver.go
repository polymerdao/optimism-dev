package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type ResolveContract interface {
	Resolve(ctx context.Context, challenge types.CommitmentArg, blob []byte) (txmgr.TxCandidate, error)
}

type ResolveContractCreator func(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (ResolveContract, error)

type Resolver struct {
	logger   log.Logger
	contract ResolveContract
	txSender types.TxSender
}

var _ Resolving = (*Resolver)(nil)

func NewResolver(l log.Logger, contract ResolveContract, txSender types.TxSender) *Resolver {
	return &Resolver{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (c *Resolver) ResolveChallenges(ctx context.Context, resolveData []types.ResolveData) (err error) {
	for _, chal := range resolveData {
		err = errors.Join(err, c.resolveChallenge(ctx, chal))
	}
	return err
}

func (c *Resolver) resolveChallenge(ctx context.Context, resolveDatum types.ResolveData) error {
	c.logger.Debug("Attempting to resolve challenge for", "blockheight", resolveDatum.ChallengedBlockNumber,
		"commitment", resolveDatum.ChallengedCommitment)

	candidate, err := c.contract.Resolve(ctx, resolveDatum.CommitmentArg, resolveDatum.Blob)
	if err != nil {
		return fmt.Errorf("failed to create candidate resolve tx: %w", err)
	}

	if err = c.txSender.SendAndWaitSimple("resolve da challenge", candidate); err != nil {
		return fmt.Errorf("failed to resolve da challenge: %w", err)
	}

	return nil
}
