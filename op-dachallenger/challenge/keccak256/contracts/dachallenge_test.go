package contracts

import (
	"context"
	"math/big"
	"math/rand"
	"slices"
	"testing"

	contractMetrics "github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching/rpcblock"
	batchingTest "github.com/ethereum-optimism/optimism/op-service/sources/batching/test"
	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/snapshots"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var (
	dacAddr        = common.HexToAddress("0x24112842371dFC380576ebb09Ae16Cb6B6caD7CB")
	balAddr        = common.HexToAddress("0x33332842371dFC380576ebb09Ae16Cb6B6c3333")
	challengerAddr = common.HexToAddress("0x44442842371dFC380576ebb09Ae16Cb6B6ca4444")
)

type contractVersion struct {
	version string
	loadAbi func() *abi.ABI
}

func (c contractVersion) Is(versions ...string) bool {
	return slices.Contains(versions, c.version)
}

const (
	versLatest = "1.0.0"
)

var (
	seed               int64 = 1337
	rnd                      = rand.New(rand.NewSource(seed))
	challengeWindow          = big.NewInt(1337)
	resolveWindow            = big.NewInt(1773)
	bondSize                 = big.NewInt(1373)
	balance                  = big.NewInt(7331)
	lockedBond               = big.NewInt(3000)
	startBlock               = big.NewInt(20000000)
	resolvedBlock            = big.NewInt(20010000)
	status                   = types.Resolved
	blob                     = make([]byte, 32*4096)
	_, _                     = rnd.Read(blob)
	challengeHeight          = big.NewInt(19999999)
	commitment               = crypto.Keccak256(blob)
	prefixedCommitment       = append([]byte{uint8(types.Keccak256)}, commitment...)
	comm                     = types.CommitmentArg{
		ChallengedBlockNumber: challengeHeight,
		ChallengedCommitment:  prefixedCommitment,
	}
	chal = &types.Challenge{
		CommitmentArg: comm,
		Challenger:    challengerAddr,
		LockedBond:    lockedBond,
		StartBlock:    startBlock,
		ResolvedBlock: resolvedBlock,
	}
)

var versions = []contractVersion{
	{
		version: versLatest,
		loadAbi: snapshots.LoadDAChallengeABI,
	},
}

func TestSimpleGetters(t *testing.T) {
	tests := []struct {
		methodAlias string
		method      string
		args        []interface{}
		result      interface{}
		expected    interface{} // Defaults to expecting the same as result
		call        func(game DAChallengeContract) (any, error)
		applies     func(version contractVersion) bool
	}{
		{
			methodAlias: "challengeWindow",
			method:      methodChallengeWindow,
			result:      uint64(1337),
			expected:    challengeWindow,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetChallengeWindow(context.Background())
			},
		},
		{
			methodAlias: "resolveWindow",
			method:      methodResolveWindow,
			result:      uint64(1773),
			expected:    resolveWindow,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetResolveWindow(context.Background())
			},
		},
		{
			methodAlias: "bondSize",
			method:      methodBondSize,
			result:      uint64(1373),
			expected:    bondSize,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetBondSize(context.Background())
			},
		},
		{
			methodAlias: "balances",
			method:      methodBalances,
			result:      uint64(7331),
			expected:    balance,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetBalance(context.Background(), balAddr)
			},
		},
		{
			methodAlias: "getChallenge",
			method:      methodGetChallenge,
			result:      chal,
			expected:    chal,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetChallenge(context.Background(), comm)
			},
		},
		{
			methodAlias: "getChallengeStatus",
			method:      methodGetChallengeStatus,
			result:      types.Resolved,
			expected:    status,
			call: func(game DAChallengeContract) (any, error) {
				return game.GetChallengeStatus(context.Background(), comm)
			},
		},
	}
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			for _, test := range tests {
				test := test
				t.Run(test.methodAlias, func(t *testing.T) {
					if test.applies != nil && !test.applies(version) {
						t.Skip("Skipping for this version")
					}
					stubRpc, game := setupDAChallengeTest(t, version)
					stubRpc.SetResponse(dacAddr, test.method, rpcblock.Latest, nil, []interface{}{test.result})
					status, err := test.call(game)
					require.NoError(t, err)
					expected := test.expected
					if expected == nil {
						expected = test.result
					}
					require.Equal(t, expected, status)
				})
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, dac := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodBalances, rpcblock.Latest, []interface{}{balAddr.Bytes()}, []interface{}{balance})

			actual, err := dac.GetBalance(context.Background(), balAddr)
			require.NoError(t, err)
			require.Equal(t, balance.Bytes(), actual.Bytes())
		})
	}
}

func TestGetChallenge(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, dac := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodGetChallenge, rpcblock.Latest,
				[]interface{}{common.BigToHash(comm.ChallengedBlockNumber).Bytes(), comm.ChallengedCommitment},
				[]interface{}{chal.Challenger, chal.LockedBond, chal.StartBlock, chal.ResolvedBlock})

			actual, err := dac.GetChallenge(context.Background(), comm)
			require.NoError(t, err)
			require.Equal(t, chal.ChallengedCommitment, chal)
			require.Equal(t, chal.Challenger, actual.Challenger)
			require.Equal(t, chal.LockedBond, actual.LockedBond)
			require.Equal(t, chal.StartBlock, actual.StartBlock)
			require.Equal(t, chal.ResolvedBlock, actual.ResolvedBlock)
		})
	}
}

func TestGetChallengeStatus(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodGetChallengeStatus, rpcblock.Latest,
				[]interface{}{common.BigToHash(comm.ChallengedBlockNumber).Bytes(), comm.ChallengedCommitment},
				[]interface{}{status.ToUint64()})
			actual, err := game.GetChallengeStatus(context.Background(), comm)
			require.NoError(t, err)
			require.Equal(t, status, actual)
		})
	}
}

func TestValidateCommitment(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodWithdraw, rpcblock.Latest, []interface{}{prefixedCommitment}, nil)
			actual, err := game.ValidateCommitment(context.Background(), prefixedCommitment)
			require.NoError(t, err)
			require.True(t, actual)
		})
	}
}

func TestComputeCommitmentKeccak256(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodComputeCommitmentKeccak256, rpcblock.Latest, []interface{}{blob},
				[]interface{}{prefixedCommitment})
			actual, err := game.ComputeCommitmentKeccak256(context.Background(), blob)
			require.NoError(t, err)
			require.Equal(t, prefixedCommitment, actual)
		})
	}
}

func TestDepositTx(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodDeposit, rpcblock.Latest, nil, nil)
			tx, err := game.Deposit(context.Background())
			require.NoError(t, err)
			stubRpc.VerifyTxCandidate(tx)
		})
	}
}

func TestWithdrawTx(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodWithdraw, rpcblock.Latest, nil, nil)
			tx, err := game.Withdraw(context.Background())
			require.NoError(t, err)
			stubRpc.VerifyTxCandidate(tx)
		})
	}
}

func TestUnlockBondTx(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodUnlockBond, rpcblock.Latest, []interface{}{common.BigToHash(comm.ChallengedBlockNumber).Bytes(),
				comm.ChallengedCommitment}, nil)
			tx, err := game.UnlockBond(context.Background(), comm)
			require.NoError(t, err)
			stubRpc.VerifyTxCandidate(tx)
		})
	}
}

func TestChallengeTx(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodChallenge, rpcblock.Latest, []interface{}{common.BigToHash(comm.ChallengedBlockNumber).Bytes(),
				comm.ChallengedCommitment}, nil)
			tx, err := game.Challenge(context.Background(), comm)
			require.NoError(t, err)
			stubRpc.VerifyTxCandidate(tx)
		})
	}
}

func TestResolveTx(t *testing.T) {
	for _, version := range versions {
		version := version
		t.Run(version.version, func(t *testing.T) {
			stubRpc, game := setupDAChallengeTest(t, version)
			stubRpc.SetResponse(dacAddr, methodResolve, rpcblock.Latest, []interface{}{common.BigToHash(comm.ChallengedBlockNumber).Bytes(),
				comm.ChallengedCommitment, blob}, nil)
			tx, err := game.Resolve(context.Background(), comm, blob)
			require.NoError(t, err)
			stubRpc.VerifyTxCandidate(tx)
		})
	}
}

func setupDAChallengeTest(t *testing.T, version contractVersion) (*batchingTest.AbiBasedRpc, DAChallengeContract) {
	dacABI := version.loadAbi()

	stubRpc := batchingTest.NewAbiBasedRpc(t, dacAddr, dacABI)
	caller := batching.NewMultiCaller(stubRpc, batching.DefaultBatchSize)

	stubRpc.SetResponse(dacAddr, methodVersion, rpcblock.Latest, nil, []interface{}{version.version})
	game, err := NewDAChallengeContract(context.Background(), contractMetrics.NoopContractMetrics, dacAddr, caller)
	require.NoError(t, err)
	return stubRpc, game
}
