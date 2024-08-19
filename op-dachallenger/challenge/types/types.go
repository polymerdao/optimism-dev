package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ChallengeStatus uint64

const (
	Uninitialized ChallengeStatus = iota
	Active
	Resolved
	Expired
)

func (cs ChallengeStatus) ToUint64() uint64 {
	return uint64(cs)
}

type CommitmentArg struct {
	ChallengedBlockNumber *big.Int
	ChallengedCommitment  []byte
}

type Challenge struct {
	CommitmentArg
	Challenger    common.Address
	LockedBond    *big.Int
	StartBlock    *big.Int
	ResolvedBlock *big.Int
}

type ResolveData struct {
	CommitmentArg
	Blob []byte
}

type Status struct {
	CommitmentArg
	Status ChallengeStatus
}
