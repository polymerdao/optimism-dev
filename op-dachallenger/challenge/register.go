package challenge

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/keccak256"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	daconfig "github.com/ethereum-optimism/optimism/op-dachallenger/config"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

type ContractCreation interface {
	ChallengeContractCreator(ctx context.Context, m metrics.ContractMetricer,
		addr common.Address, caller *batching.MultiCaller) (types.ChallengeContract, error)
	WithdrawContractCreator(ctx context.Context, m metrics.ContractMetricer,
		addr common.Address, caller *batching.MultiCaller) (types.WithdrawContract, error)
	ResolveContractCreator(ctx context.Context, m metrics.ContractMetricer,
		addr common.Address, caller *batching.MultiCaller) (types.ResolveContract, error)
	UnlockContractCreator(ctx context.Context, m metrics.ContractMetricer,
		addr common.Address, caller *batching.MultiCaller) (types.UnlockContract, error)
}

func NewContractCreator(daType plasma.CommitmentType, daKind daconfig.CommitmentKind) (ContractCreation, error) {
	switch daType {
	case plasma.Keccak256CommitmentType:
		return &keccak256.ContractCreator{}, nil
	case plasma.GenericCommitmentType:
		switch daKind {
		case daconfig.Undefined:
			return nil, errors.New("daKind is undefined")
		case daconfig.EigenDA:
			return nil, errors.New("daKind EigenDA is currently unsupported")
		default:
			return nil, fmt.Errorf("unrecongized daKind: %d", daKind)
		}
	default:
		return nil, fmt.Errorf("unrecongized daType: %d", daType)
	}
}

type Registration interface {
	RegisterBondWithdrawContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
		creator scheduler.WithdrawContractCreator)
	RegisterCommitmentChallengeContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
		creator scheduler.ChallengeContractCreator)
	RegisterBondUnlockContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
		creator scheduler.UnlockerContractCreator)
	RegisterChallengeResolverContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
		creator scheduler.ResolveContractCreator)
}

var _ Registration = &Registry{}

type Registry struct {
	CreateChallengeContract map[plasma.CommitmentType]map[daconfig.CommitmentKind]func(ctx context.Context,
		m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) (types.ChallengeContract, error)
	CreateUnlockContract map[plasma.CommitmentType]map[daconfig.CommitmentKind]func(ctx context.Context,
		m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) (types.UnlockContract, error)
	CreateWithdrawContract map[plasma.CommitmentType]map[daconfig.CommitmentKind]func(ctx context.Context,
		m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) (types.WithdrawContract, error)
	CreateResolveContract map[plasma.CommitmentType]map[daconfig.CommitmentKind]func(ctx context.Context,
		m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) (types.ResolveContract, error)
}

func (r *Registry) RegisterBondWithdrawContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
	creator scheduler.WithdrawContractCreator) {
	r.CreateWithdrawContract[daType][daKind] = creator
}

func (r *Registry) RegisterCommitmentChallengeContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
	creator scheduler.ChallengeContractCreator) {
	r.CreateChallengeContract[daType][daKind] = creator
}

func (r *Registry) RegisterBondUnlockContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
	creator scheduler.UnlockerContractCreator) {
	r.CreateUnlockContract[daType][daKind] = creator
}

func (r *Registry) RegisterChallengeResolverContract(daType plasma.CommitmentType, daKind daconfig.CommitmentKind,
	creator scheduler.ResolveContractCreator) {
	r.CreateResolveContract[daType][daKind] = creator
}
