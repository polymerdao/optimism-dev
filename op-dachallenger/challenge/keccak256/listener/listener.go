package listener

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// Listener for
// 1. New commitments submitted to the batch inbox
// 2. New Challenge events emitted from the DAChallenge contract
// 3. New Resolve events emitted from the DAChallenge contract
// 4. New Head events indicating a new block has been added to the L1 chain
// and funnel the events into a single out channel for the solver to figure out what to do
type Listener struct {
	inbox      <-chan types.CommitmentArg
	challenges <-chan types.Challenge
	statuses   <-chan types.Status
	heads      <-chan ethTypes.Header
	headsMod   *big.Int // only trigger new Heads event at some interval
	out        chan<- Event
	quit       <-chan struct{}
}

type EventType uint64

const (
	InboxEvent EventType = iota
	ChallengeEvent
	StatusEvent
	NewHead
)

type Event struct {
	Type           EventType
	InboxEvent     types.CommitmentArg
	ChallengeEvent types.Challenge
	StatusEvent    types.Status
	Header         ethTypes.Header
}

// Source for Listener
type Source interface {
	InboxEvents() (<-chan types.CommitmentArg, error)
	ChallengeEvents() (<-chan types.Challenge, error)
	StatusEvents() (<-chan types.Status, error)
	NewHeads() (<-chan ethTypes.Header, error)
}

// NewListener creates a new Listener from the given Source
func NewListener(source Source, headsMod int64, out chan<- Event, quit <-chan struct{}) (*Listener, error) {
	if headsMod <= 0 {
		headsMod = 1
	}
	inbox, err := source.InboxEvents()
	if err != nil {
		return nil, err
	}
	challenges, err := source.ChallengeEvents()
	if err != nil {
		return nil, err
	}
	statuses, err := source.StatusEvents()
	if err != nil {
		return nil, err
	}
	heads, err := source.NewHeads()
	if err != nil {
		return nil, err
	}
	return &Listener{
		inbox:      inbox,
		challenges: challenges,
		statuses:   statuses,
		heads:      heads,
		headsMod:   big.NewInt(headsMod),
		out:        out,
		quit:       quit,
	}, nil
}

func (l *Listener) Listen() {
	for {
		select {
		case <-l.quit:
			return
		case inbox := <-l.inbox:
			l.out <- Event{
				Type:       InboxEvent,
				InboxEvent: inbox,
			}
		case challenge := <-l.challenges:
			l.out <- Event{
				Type:           ChallengeEvent,
				ChallengeEvent: challenge,
			}
		case status := <-l.statuses:
			l.out <- Event{
				Type:        StatusEvent,
				StatusEvent: status,
			}
		case head := <-l.heads:
			if head.Number.Mod(head.Number, l.headsMod).Cmp(big.NewInt(0)) == 0 {
				l.out <- Event{
					Type:   NewHead,
					Header: head,
				}
			}
		}
	}
}
