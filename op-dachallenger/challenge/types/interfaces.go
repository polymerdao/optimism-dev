package types

import (
	"context"
	"io"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

type PlasmaFetcher interface {
	derive.PlasmaInputFetcher
	GetChallengeStatus(comm plasma.CommitmentData, commBlockNumber uint64) plasma.ChallengeStatus
}

type TxSender interface {
	SendAndWaitSimple(txPurpose string, txs ...txmgr.TxCandidate) error
}

type Actor interface {
	OnNewL1Finalized(ctx context.Context, finalized eth.L1BlockRef)
	Start(ctx context.Context)
	io.Closer
}

type ChallengeContract interface {
	Challenge(ctx context.Context, challenge CommitmentArg) (txmgr.TxCandidate, error)
}

type ResolveContract interface {
	Resolve(ctx context.Context, challenge CommitmentArg, blob []byte) (txmgr.TxCandidate, error)
}

type UnlockContract interface {
	UnlockBond(ctx context.Context, challenge CommitmentArg) (txmgr.TxCandidate, error)
}

type WithdrawContract interface {
	Withdraw(ctx context.Context) (txmgr.TxCandidate, error)
}
