package poly_engineapi

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/polymerdao/monomer/builder"
	monoengine "github.com/polymerdao/monomer/engine"
)

// L2EngineAPI wraps a monomer EngineAPI, and implements the RPC backend required to serve the engine API.
// This re-implements some of the Geth API work, but changes the API backend so we can deterministically
// build and control the L2 block contents to reach very specific edge cases as desired for testing.
type L2EngineAPI struct {
	log     log.Logger
	backend *monoengine.EngineAPI

	// Functionality for snap sync
	remotes map[common.Hash]*types.Block

	// L2 block building data
	pendingIndices map[common.Address]uint64 // per account, how many txs from the pool were already included in the block, since the pool is lagging behind block mining.
	l2ForceEmpty   bool                      // when no additional txs may be processed (i.e. when sequencer drift runs out)
	l2TxFailed     []*types.Transaction      // log of failed transactions which could not be included
}

func NewL2EngineAPI(log log.Logger, b *builder.Builder,
	txValidator monoengine.TxValidator,
	blockStore monoengine.BlockStore) *L2EngineAPI {
	return &L2EngineAPI{
		log:     log,
		backend: monoengine.NewEngineAPI(b, txValidator, blockStore),
		remotes: make(map[common.Hash]*types.Block),
	}
}

// computePayloadId computes a pseudo-random payloadid, based on the parameters.
func computePayloadId(headBlockHash common.Hash, attrs *eth.PayloadAttributes) engine.PayloadID {
	// Hash
	hasher := sha256.New()
	hasher.Write(headBlockHash[:])
	_ = binary.Write(hasher, binary.BigEndian, attrs.Timestamp)
	hasher.Write(attrs.PrevRandao[:])
	hasher.Write(attrs.SuggestedFeeRecipient[:])
	_ = binary.Write(hasher, binary.BigEndian, attrs.NoTxPool)
	_ = binary.Write(hasher, binary.BigEndian, uint64(len(attrs.Transactions)))
	for _, tx := range attrs.Transactions {
		_ = binary.Write(hasher, binary.BigEndian, uint64(len(tx))) // length-prefix to avoid collisions
		hasher.Write(tx)
	}
	_ = binary.Write(hasher, binary.BigEndian, *attrs.GasLimit)
	var out engine.PayloadID
	copy(out[:], hasher.Sum(nil)[:8])
	return out
}

func (ea *L2EngineAPI) RemainingBlockGas() uint64 {
	panic("implement me") // this is only used in certain tests
}

func (ea *L2EngineAPI) ForcedEmpty() bool {
	return ea.l2ForceEmpty
}

func (ea *L2EngineAPI) PendingIndices(from common.Address) uint64 {
	return ea.pendingIndices[from]
}

var ErrNotBuildingBlock = errors.New("not currently building a block, cannot include tx from queue")

func (ea *L2EngineAPI) IncludeTx(tx *types.Transaction, from common.Address) error {
	panic("implement me") // this is only used in some tests
}

func (ea *L2EngineAPI) GetPayloadV1(ctx context.Context, payloadId eth.PayloadID) (*eth.ExecutionPayload, error) {
	env, err := ea.backend.GetPayloadV1(ctx, payloadId)
	if err != nil {
		return nil, err
	}
	return env.ExecutionPayload, nil
}

func (ea *L2EngineAPI) GetPayloadV2(ctx context.Context, payloadId eth.PayloadID) (*eth.ExecutionPayloadEnvelope, error) {
	return ea.backend.GetPayloadV2(ctx, payloadId)
}

func (ea *L2EngineAPI) GetPayloadV3(ctx context.Context, payloadId eth.PayloadID) (*eth.ExecutionPayloadEnvelope, error) {
	return ea.backend.GetPayloadV3(ctx, payloadId)
}

func (ea *L2EngineAPI) ForkchoiceUpdatedV1(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	return ea.backend.ForkchoiceUpdatedV1(ctx, *state, attr)
}

func (ea *L2EngineAPI) ForkchoiceUpdatedV2(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	return ea.backend.ForkchoiceUpdatedV2(ctx, *state, attr)
}

// Ported from: https://github.com/ethereum-optimism/op-geth/blob/c50337a60a1309a0f1dca3bf33ed1bb38c46cdd7/eth/catalyst/api.go#L197C1-L205C1
func (ea *L2EngineAPI) ForkchoiceUpdatedV3(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	return ea.backend.ForkchoiceUpdatedV3(ctx, *state, attr)

}

func checkAttribute(active func(*big.Int, uint64) bool, exists bool, block *big.Int, time uint64) error {
	if active(block, time) && !exists {
		return errors.New("fork active, missing expected attribute")
	}
	if !active(block, time) && exists {
		return errors.New("fork inactive, unexpected attribute set")
	}
	return nil
}

func (ea *L2EngineAPI) NewPayloadV1(ctx context.Context, payload *eth.ExecutionPayload) (*eth.PayloadStatusV1, error) {
	if payload.Withdrawals != nil {
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, engine.InvalidParams.With(errors.New("withdrawals not supported in V1"))
	}

	return ea.backend.NewPayloadV1(*payload)
}

func (ea *L2EngineAPI) NewPayloadV2(ctx context.Context, payload *eth.ExecutionPayload) (*eth.PayloadStatusV1, error) {
	return ea.backend.NewPayloadV2(*payload)
}

// Ported from: https://github.com/ethereum-optimism/op-geth/blob/c50337a60a1309a0f1dca3bf33ed1bb38c46cdd7/eth/catalyst/api.go#L486C1-L507
func (ea *L2EngineAPI) NewPayloadV3(ctx context.Context, params *eth.ExecutionPayload, versionedHashes []common.Hash, beaconRoot *common.Hash) (*eth.PayloadStatusV1, error) {
	if params.ExcessBlobGas == nil {
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, engine.InvalidParams.With(errors.New("nil excessBlobGas post-cancun"))
	}
	if params.BlobGasUsed == nil {
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, engine.InvalidParams.With(errors.New("nil params.BlobGasUsed post-cancun"))
	}
	if versionedHashes == nil {
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, engine.InvalidParams.With(errors.New("nil versionedHashes post-cancun"))
	}
	if beaconRoot == nil {
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, engine.InvalidParams.With(errors.New("nil parentBeaconBlockRoot post-cancun"))
	}

	return ea.backend.NewPayloadV3(*params)
}
