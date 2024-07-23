package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/log"
)

type BondContract interface {
	UnlockBond(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error)
}

type Unlocker struct {
	logger   log.Logger
	contract BondContract
	txSender TxSender
}

var _ BondUnlocker = (*Unlocker)(nil)

func NewBondUnlocker(l log.Logger, contract BondContract, txSender TxSender) *Unlocker {
	return &Unlocker{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (c *Unlocker) UnlockBonds(ctx context.Context, challenges []types.CommitmentArg) (err error) {
	for _, chal := range challenges {
		err = errors.Join(err, c.unlockBond(ctx, chal))
	}
	return err
}

func (c *Unlocker) unlockBond(ctx context.Context, challenge types.CommitmentArg) error {
	c.logger.Debug("Attempting to unlock bonds for", "blockheight", challenge.ChallengedBlockNumber,
		"commitment", challenge.ChallengedCommitment)

	candidate, err := c.contract.UnlockBond(ctx, challenge)
	if errors.Is(err, contracts.ErrSimulationFailed) {
		c.logger.Debug("Bond is still locked", "height", challenge.ChallengedBlockNumber,
			"commitment", challenge.ChallengedCommitment)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to create candidate unlock bond tx: %w", err)
	}

	if err = c.txSender.SendAndWaitSimple("unlock bond", candidate); err != nil {
		return fmt.Errorf("failed to unlock bond: %w", err)
	}

	return nil
}
