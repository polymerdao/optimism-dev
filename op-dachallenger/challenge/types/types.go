package types

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
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

type CommitmentType int

const (
	Keccak256 CommitmentType = iota
)

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

type ActionType string

func (a ActionType) String() string {
	return string(a)
}

const (
	ActionChallenge ActionType = "challenge"
	ActionResolve   ActionType = "resolve"
)

type Action struct {
	Type ActionType

	// Existing challenge, if this is a resolve action
	Challenge Challenge

	// Resolve data, if this is a challenge action
	ResolveData ResolveData
}

type TxSender interface {
	SendAndWaitSimple(txPurpose string, txs ...txmgr.TxCandidate) error
}
