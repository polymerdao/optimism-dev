package contracts

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching/rpcblock"
	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/snapshots"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	methodVersion                    = "version"
	methodChallengeWindow            = "challengeWindow"
	methodResolveWindow              = "resolveWindow"
	methodBondSize                   = "bondSize"
	methodBalances                   = "balances"
	methodDeposit                    = "deposit"
	methodWithdraw                   = "withdraw"
	methodGetChallenge               = "getChallenge"
	methodGetChallengeStatus         = "getChallengeStatus"
	methodValidateCommitment         = "validateCommitment"
	methodChallenge                  = "challenge"
	methodResolve                    = "resolve"
	methodUnlockBond                 = "unlockBond"
	methodComputeCommitmentKeccak256 = "computeCommitmentKeccak256"
)

type DAChallengeContract interface {
	GetChallengeWindow()
	GetResolveWindow()
	GetBondSize()
	GetChallenge()
	GetChallengeStatus()
	GetBalance()
	Deposit()
	Challenge()
	UnlockBond()
	Withdraw()
	ValidateCommitment()
	ComputeCommitmentKeccak256()
	Resolve()
}

var _ DAChallengeContract = &DAChallengeContractLatest{}

type DAChallengeContractLatest struct {
	metrics     metrics.ContractMetricer
	multiCaller *batching.MultiCaller
	contract    *batching.BoundContract
	abi         *abi.ABI
}

type Challenge struct {
	Challenger    common.Address
	LockedBond    *big.Int
	StartBlock    *big.Int
	ResolvedBlock *big.Int
}

type challenge struct {
	Challenger    [32]byte
	LockedBond    [32]byte
	StartBlock    [32]byte
	ResolvedBlock [32]byte
}

type CommitmentStatus int

const (
	Uninitialized CommitmentStatus = iota
	Active
	Resolved
	Expired
)

type CommitmentType int

const (
	Keccak256 CommitmentType = iota
)

func NewDAChallengeContract(ctx context.Context, m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) (DAChallengeContract, error) {
	contractAbi := snapshots.LoadDAChallengeABI()

	result, err := caller.SingleCall(ctx, rpcblock.Latest, batching.NewContractCall(contractAbi, addr, methodVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve version of dispute game %v: %w", addr, err)
	}
	version := result.GetString(0)
	supportedVersion := "1.0.0"
	if version != supportedVersion {
		return nil, fmt.Errorf("unrecognized version for dachallenge contract %s, expected %s", version, supportedVersion)
	} else {
		return &DAChallengeContractLatest{
			metrics:     m,
			multiCaller: caller,
			contract:    batching.NewBoundContract(contractAbi, addr),
			abi:         contractAbi,
		}, nil
	}
}

func (D DAChallengeContractLatest) GetChallengeWindow() {
	panic("implement me")
}

func (D DAChallengeContractLatest) GetResolveWindow() {
	panic("implement me")
}

func (D DAChallengeContractLatest) GetBondSize() {
	panic("implement me")
}

func (D DAChallengeContractLatest) GetChallenge() {
	panic("implement me")
}

func (D DAChallengeContractLatest) GetChallengeStatus() {
	panic("implement me")
}

func (D DAChallengeContractLatest) GetBalance() {
	panic("implement me")
}

func (D DAChallengeContractLatest) Deposit() {
	panic("implement me")
}

func (D DAChallengeContractLatest) Challenge() {
	panic("implement me")
}

func (D DAChallengeContractLatest) UnlockBond() {
	panic("implement me")
}

func (D DAChallengeContractLatest) Withdraw() {
	panic("implement me")
}

func (D DAChallengeContractLatest) ValidateCommitment() {
	panic("implement me")
}

func (D DAChallengeContractLatest) ComputeCommitmentKeccak256() {
	panic("implement me")
}

func (D DAChallengeContractLatest) Resolve() {
	panic("implement me")
}
