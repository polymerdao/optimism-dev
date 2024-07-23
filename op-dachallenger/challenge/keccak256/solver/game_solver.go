package solver

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/types"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum/go-ethereum/common"
)

type GameSolver struct {
	claimSolver *claimSolver
}

func NewGameSolver(client plasma.DAClient) *GameSolver {
	return &GameSolver{
		claimSolver: newClaimSolver(client),
	}
}

func (s *GameSolver) CalculateNextActions(ctx context.Context, game types.Game) ([]types.Action, error) {
	var actions []types.Action
	agreedClaims := newHonestClaimTracker()
	if agreeWithRootClaim {
		agreedClaims.AddHonestClaim(types.Claim{}, game.Claims()[0])
	}
	for _, claim := range game.Claims() {
		var action *types.Action
		if claim.Depth() == game.MaxDepth() {
			action, err = s.calculateStep(ctx, game, claim, agreedClaims)
		} else {
			action, err = s.calculateMove(ctx, game, claim, agreedClaims)
		}
		if err != nil {
			// Unable to continue iterating claims safely because we may not have tracked the required honest moves
			// for this claim which affects the response to later claims.
			// Any actions we've already identified are still safe to apply.
			return actions, fmt.Errorf("failed to determine response to claim %v: %w", claim.ContractIndex, err)
		}
		if action == nil {
			continue
		}
		actions = append(actions, *action)
	}
	return actions, nil
}

func (s *GameSolver) calculateStep(ctx context.Context, game types.Game, claim types.Claim, agreedClaims *honestClaimTracker) (*types.Action, error) {
	if claim.CounteredBy != (common.Address{}) {
		return nil, nil
	}
	step, err := s.claimSolver.AttemptStep(ctx, game, claim, agreedClaims)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, nil
	}
	return &types.Action{
		Type:        types.ActionTypeStep,
		ParentClaim: step.LeafClaim,
		IsAttack:    step.IsAttack,
		PreState:    step.PreState,
		ProofData:   step.ProofData,
		OracleData:  step.OracleData,
	}, nil
}

func (s *GameSolver) calculateMove(ctx context.Context, game types.Game, claim types.Claim, honestClaims *honestClaimTracker) (*types.Action, error) {
	move, err := s.claimSolver.NextMove(ctx, claim, game, honestClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next move for claim index %v: %w", claim.ContractIndex, err)
	}
	if move == nil {
		return nil, nil
	}
	honestClaims.AddHonestClaim(claim, *move)
	if game.IsDuplicate(*move) {
		return nil, nil
	}
	return &types.Action{
		Type:        types.ActionTypeMove,
		IsAttack:    !game.DefendsParent(*move),
		ParentClaim: game.Claims()[move.ParentContractIndex],
		Value:       move.Value,
	}, nil
}
