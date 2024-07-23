package contracts

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching/rpcblock"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
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
	GetChallengeWindow(ctx context.Context) (*big.Int, error)
	GetResolveWindow(ctx context.Context) (*big.Int, error)
	GetBondSize(ctx context.Context) (*big.Int, error)
	GetChallenge(ctx context.Context, challenge types.CommitmentArg) (*types.Challenge, error)
	GetChallengeStatus(ctx context.Context, challenge types.CommitmentArg) (types.ChallengeStatus, error)
	GetBalance(ctx context.Context, addr common.Address) (*big.Int, error)
	Deposit(ctx context.Context) (txmgr.TxCandidate, error)
	Challenge(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error)
	UnlockBond(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error)
	Withdraw(ctx context.Context) (txmgr.TxCandidate, error)
	ValidateCommitment(ctx context.Context, commitment []byte) (bool, error)
	ComputeCommitmentKeccak256(ctx context.Context, blob []byte) ([]byte, error)
	Resolve(ctx context.Context, challenge types.CommitmentArg, blob []byte) (txmgr.TxCandidate, error)
}

var _ DAChallengeContract = &DAChallengeContractLatest{}

type DAChallengeContractLatest struct {
	metrics     metrics.ContractMetricer
	multiCaller *batching.MultiCaller
	contract    *batching.BoundContract
	abi         *abi.ABI
}

type challenge struct {
	Challenger    [32]byte
	LockedBond    [32]byte
	StartBlock    [32]byte
	ResolvedBlock [32]byte
}

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

func (d *DAChallengeContractLatest) GetChallengeWindow(ctx context.Context) (*big.Int, error) {
	defer d.metrics.StartContractRequest("GetChallengeWindow")()
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodChallengeWindow))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve challenge window: %w", err)
	}
	return res.GetBigInt(0), nil
}

func (d *DAChallengeContractLatest) GetResolveWindow(ctx context.Context) (*big.Int, error) {
	defer d.metrics.StartContractRequest("GetResolveWindow")()
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodResolveWindow))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the resolve window: %w", err)
	}
	return res.GetBigInt(0), nil
}

func (d *DAChallengeContractLatest) GetBondSize(ctx context.Context) (*big.Int, error) {
	defer d.metrics.StartContractRequest("GetBondSize")()
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodBondSize))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the bond size: %w", err)
	}
	return res.GetBigInt(0), nil
}

func (d *DAChallengeContractLatest) GetChallenge(ctx context.Context, challenge types.CommitmentArg) (*types.Challenge, error) {
	defer d.metrics.StartContractRequest("GetChallenge")()
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodGetChallenge,
		common.BigToHash(challenge.ChallengedBlockNumber).Bytes(), challenge.ChallengedCommitment))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the challenge: %w", err)
	}
	return &types.Challenge{
		CommitmentArg: challenge,
		Challenger:    res.GetAddress(0),
		LockedBond:    res.GetBigInt(1),
		StartBlock:    res.GetBigInt(2),
		ResolvedBlock: res.GetBigInt(3),
	}, nil
}

func (d *DAChallengeContractLatest) GetChallengeStatus(ctx context.Context, challenge types.CommitmentArg) (types.ChallengeStatus, error) {
	defer d.metrics.StartContractRequest("GetChallengeStatus")
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodGetChallengeStatus,
		common.BigToHash(challenge.ChallengedBlockNumber).Bytes(), challenge.ChallengedCommitment))
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve the challenge status: %w", err)
	}
	return types.ChallengeStatus(res.GetUint64(0)), nil
}

func (d *DAChallengeContractLatest) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	defer d.metrics.StartContractRequest("GetBalance")
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodBalances,
		addr.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the balance: %w", err)
	}
	return res.GetBigInt(0), nil
}

func (d *DAChallengeContractLatest) Deposit(ctx context.Context) (txmgr.TxCandidate, error) {
	defer d.metrics.StartContractRequest("Deposit")
	tx, err := d.contract.Call(methodDeposit).ToTxCandidate()
	if err != nil {
		return txmgr.TxCandidate{}, fmt.Errorf("failed to create deposit tx: %w", err)
	}
	return tx, nil
}

func (d *DAChallengeContractLatest) Challenge(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error) {
	defer d.metrics.StartContractRequest("Challenge")
	tx, err := d.contract.Call(methodChallenge, common.BigToHash(challenge.ChallengedBlockNumber).Bytes(),
		challenge.ChallengedCommitment).ToTxCandidate()
	if err != nil {
		return txmgr.TxCandidate{}, fmt.Errorf("failed to create challenge tx: %w", err)
	}
	return tx, nil
}

func (d *DAChallengeContractLatest) UnlockBond(ctx context.Context, challenge types.CommitmentArg) (txmgr.TxCandidate, error) {
	defer d.metrics.StartContractRequest("UnlockBond")
	tx, err := d.contract.Call(methodUnlockBond, common.BigToHash(challenge.ChallengedBlockNumber).Bytes(),
		challenge.ChallengedCommitment).ToTxCandidate()
	if err != nil {
		return txmgr.TxCandidate{}, fmt.Errorf("failed to create unlock bond tx: %w", err)
	}
	return tx, nil
}

func (d *DAChallengeContractLatest) Withdraw(ctx context.Context) (txmgr.TxCandidate, error) {
	defer d.metrics.StartContractRequest("Withdraw")
	tx, err := d.contract.Call(methodWithdraw).ToTxCandidate()
	if err != nil {
		return txmgr.TxCandidate{}, fmt.Errorf("failed to create withdraw tx: %w", err)
	}
	return tx, nil
}

func (d *DAChallengeContractLatest) ValidateCommitment(ctx context.Context, commitment []byte) (bool, error) {
	defer d.metrics.StartContractRequest("ValidateCommitment")
	_, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodValidateCommitment, commitment))
	if err != nil {
		return false, fmt.Errorf("failed to validate commitment: %w", err)
	}
	return true, nil
}

func (d *DAChallengeContractLatest) ComputeCommitmentKeccak256(ctx context.Context, blob []byte) ([]byte, error) {
	defer d.metrics.StartContractRequest("ComputeCommitmentKeccak256")
	res, err := d.multiCaller.SingleCall(ctx, rpcblock.Latest, d.contract.Call(methodComputeCommitmentKeccak256, blob))
	if err != nil {
		return nil, fmt.Errorf("failed to compute commitment keccak256: %w", err)
	}
	return res.GetBytes(0), nil
}

func (d *DAChallengeContractLatest) Resolve(ctx context.Context, challenge types.CommitmentArg, blob []byte) (txmgr.TxCandidate, error) {
	defer d.metrics.StartContractRequest("Resolve")
	tx, err := d.contract.Call(methodResolve, common.BigToHash(challenge.ChallengedBlockNumber).Bytes(),
		challenge.ChallengedCommitment, blob).ToTxCandidate()
	if err != nil {
		return txmgr.TxCandidate{}, fmt.Errorf("failed to create withdraw tx: %w", err)
	}
	return tx, nil
}
