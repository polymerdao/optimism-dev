package l2

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-program/client/l2/poly_engineapi"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/polymerdao/monomer/builder"
	monoengine "github.com/polymerdao/monomer/engine"
	monoeth "github.com/polymerdao/monomer/eth"
)

type PolyOracleEngine struct {
	api           *poly_engineapi.L2EngineAPI
	rollupCfg     *rollup.Config
	ethBlockstore *monoeth.Block
}

// once we have our oracle-backed poly_engine, use that instead of passing in the builder, txValidator, and blockstore here
func NewPolyOracleEngine(rollupCfg *rollup.Config, logger log.Logger, b *builder.Builder,
	txValidator monoengine.TxValidator,
	blockStore monoengine.BlockStore) *PolyOracleEngine {
	engineAPI := poly_engineapi.NewL2EngineAPI(logger, b, txValidator, blockStore)
	return &PolyOracleEngine{
		api:           engineAPI,
		rollupCfg:     rollupCfg,
		ethBlockstore: monoeth.NewBlock(blockStore),
	}
}

func (o *PolyOracleEngine) L2OutputRoot(l2ClaimBlockNum uint64) (eth.Bytes32, error) {
	ethHeader, err := o.ethBlockstore.GetEthHeaderByNumber(monoeth.BlockID{Height: int64(l2ClaimBlockNum), Label: ""})
	if err != nil {
		return eth.Bytes32{}, err
	}
	if ethHeader == nil {
		return eth.Bytes32{}, fmt.Errorf("failed to get L2 block at %d", l2ClaimBlockNum)
	}
	return rollup.ComputeL2OutputRootV0(eth.HeaderBlockInfo(ethHeader), [32]byte{})
}

func (o *PolyOracleEngine) GetPayload(ctx context.Context, payloadInfo eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error) {
	var res *eth.ExecutionPayloadEnvelope
	var err error
	switch method := o.rollupCfg.GetPayloadVersion(payloadInfo.Timestamp); method {
	case eth.GetPayloadV3:
		res, err = o.api.GetPayloadV3(ctx, payloadInfo.ID)
	case eth.GetPayloadV2:
		res, err = o.api.GetPayloadV2(ctx, payloadInfo.ID)
	default:
		return nil, fmt.Errorf("unsupported GetPayload version: %s", method)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *PolyOracleEngine) ForkchoiceUpdate(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	switch method := o.rollupCfg.ForkchoiceUpdatedVersion(attr); method {
	case eth.FCUV3:
		return o.api.ForkchoiceUpdatedV3(ctx, state, attr)
	case eth.FCUV2:
		return o.api.ForkchoiceUpdatedV2(ctx, state, attr)
	case eth.FCUV1:
		return o.api.ForkchoiceUpdatedV1(ctx, state, attr)
	default:
		return nil, fmt.Errorf("unsupported ForkchoiceUpdated version: %s", method)
	}
}

func (o *PolyOracleEngine) NewPayload(ctx context.Context, payload *eth.ExecutionPayload, parentBeaconBlockRoot *common.Hash) (*eth.PayloadStatusV1, error) {
	switch method := o.rollupCfg.NewPayloadVersion(uint64(payload.Timestamp)); method {
	case eth.NewPayloadV3:
		return o.api.NewPayloadV3(ctx, payload, []common.Hash{}, parentBeaconBlockRoot)
	case eth.NewPayloadV2:
		return o.api.NewPayloadV2(ctx, payload)
	default:
		return nil, fmt.Errorf("unsupported NewPayload version: %s", method)
	}
}

func (o *PolyOracleEngine) PayloadByHash(ctx context.Context, hash common.Hash) (*eth.ExecutionPayloadEnvelope, error) {
	block, err := o.ethBlockstore.GetEthBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, ErrNotFound
	}
	return eth.BlockAsPayloadEnv(block, o.rollupCfg.CanyonTime)
}

func (o *PolyOracleEngine) PayloadByNumber(ctx context.Context, n uint64) (*eth.ExecutionPayloadEnvelope, error) {
	block, err := o.ethBlockstore.GetEthBlockByNumber(monoeth.BlockID{Height: int64(n), Label: ""})
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, ErrNotFound
	}
	return eth.BlockAsPayloadEnv(block, o.rollupCfg.CanyonTime)
}

func (o *PolyOracleEngine) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	block, err := o.ethBlockstore.GetEthBlockByNumber(monoeth.BlockID{Height: 0, Label: label})
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	if block == nil {
		return eth.L2BlockRef{}, ErrNotFound
	}
	return derive.L2BlockToBlockRef(o.rollupCfg, block)
}

func (o *PolyOracleEngine) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := o.ethBlockstore.GetEthBlockByHash(l2Hash)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	if block == nil {
		return eth.L2BlockRef{}, ErrNotFound
	}
	return derive.L2BlockToBlockRef(o.rollupCfg, block)
}

func (o *PolyOracleEngine) L2BlockRefByNumber(ctx context.Context, n uint64) (eth.L2BlockRef, error) {
	block, err := o.ethBlockstore.GetEthBlockByNumber(monoeth.BlockID{Height: int64(n), Label: ""})
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	if block == nil {
		return eth.L2BlockRef{}, ErrNotFound
	}
	return derive.L2BlockToBlockRef(o.rollupCfg, block)
}

func (o *PolyOracleEngine) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	payload, err := o.PayloadByHash(ctx, hash)
	if err != nil {
		return eth.SystemConfig{}, err
	}
	return derive.PayloadToSystemConfig(o.rollupCfg, payload.ExecutionPayload)
}
