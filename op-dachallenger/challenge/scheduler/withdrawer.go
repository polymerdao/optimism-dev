package scheduler

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type WithdrawContract interface {
	Withdraw(ctx context.Context) (txmgr.TxCandidate, error)
}

type WithdrawContractCreator func(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (WithdrawContract, error)

type Withdrawer struct {
	logger   log.Logger
	contract WithdrawContract
	txSender types.TxSender
}

var _ BondWithdrawer = (*Withdrawer)(nil)

func NewBondWithdrawer(l log.Logger, contract WithdrawContract, txSender types.TxSender) *Withdrawer {
	return &Withdrawer{
		logger:   l,
		contract: contract,
		txSender: txSender,
	}
}

func (w *Withdrawer) WithdrawBonds(ctx context.Context, height uint64) (err error) {
	w.logger.Debug("Attempting to withdraw bonds for", "blockheight", height)

	candidate, err := w.contract.Withdraw(ctx)
	if err != nil {
		return fmt.Errorf("failed to create candidate withdraw tx: %w", err)
	}

	if err = w.txSender.SendAndWaitSimple("withdraw bonds", candidate); err != nil {
		return fmt.Errorf("failed to withdraw bonds: %w", err)
	}

	return nil
}
