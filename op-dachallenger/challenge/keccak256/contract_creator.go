package keccak256

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/keccak256/contracts"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum/common"
)

type ContractCreator struct{}

func (c *ContractCreator) ChallengeContractCreator(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (scheduler.ChallengeContract, error) {
	return contracts.NewDAChallengeContract(ctx, m, addr, caller)
}

func (c *ContractCreator) WithdrawContractCreator(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (scheduler.WithdrawContract, error) {
	return contracts.NewDAChallengeContract(ctx, m, addr, caller)
}

func (c *ContractCreator) ResolveContractCreator(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (scheduler.ResolveContract, error) {
	return contracts.NewDAChallengeContract(ctx, m, addr, caller)
}

func (c *ContractCreator) UnlockContractCreator(ctx context.Context, m metrics.ContractMetricer,
	addr common.Address, caller *batching.MultiCaller) (scheduler.UnlockContract, error) {
	return contracts.NewDAChallengeContract(ctx, m, addr, caller)
}
