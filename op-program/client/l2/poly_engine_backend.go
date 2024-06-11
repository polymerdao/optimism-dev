package l2

import (
	"fmt"
	"github.com/polymerdao/monomer/builder"
	monoengine "github.com/polymerdao/monomer/engine"
	"github.com/polymerdao/monomer/mempool"
	"math/big"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/ethereum-optimism/optimism/op-program/client/l2/poly_engineapi"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type OracleBackedPolymeraseChain struct {
	builder     *builder.Builder
	txValidator monoengine.TxValidator
	blockStore  monoengine.BlockStore
	log         log.Logger
	oracle      Oracle
	chainCfg    *params.ChainConfig
	engine      consensus.Engine
	oracleHead  *types.Header
	head        *types.Header
	safe        *types.Header
	finalized   *types.Header

	// Block by number cache
	hashByNum            map[uint64]common.Hash
	earliestIndexedBlock *types.Header

	// Inserted blocks
	blocks map[common.Hash]*types.Block
	db     dbm.DB
}

var _ poly_engineapi.EngineBackend = (*OracleBackedPolymeraseChain)(nil)

func NewOracleBackedPolymeraseChain(logger log.Logger, oracle Oracle, chainCfg *params.ChainConfig, l2OutputRoot common.Hash) (*OracleBackedPolymeraseChain, error) {
	// create oracle-backed versions of builder, txValidator, and blockStore
	tmdb := NewOracleBackedIAVLDB(oracle, nil)
	builder := builder.New(tmdb)
	mem := mempool.New(tmdb)

	output := oracle.OutputByRoot(l2OutputRoot)
	outputV0, ok := output.(*eth.OutputV0)
	if !ok {
		return nil, fmt.Errorf("unsupported L2 output version: %d", output.Version())
	}
	head := oracle.BlockByHash(outputV0.BlockHash)
	logger.Info("Loaded L2 head", "hash", head.Hash(), "number", head.Number())
	return &OracleBackedPolymeraseChain{
		log:      logger,
		oracle:   oracle,
		chainCfg: chainCfg,
		engine:   beacon.New(nil),

		hashByNum: map[uint64]common.Hash{
			head.NumberU64(): head.Hash(),
		},
		earliestIndexedBlock: head.Header(),

		// Treat the agreed starting head as finalized - nothing before it can be disputed
		head:       head.Header(),
		safe:       head.Header(),
		finalized:  head.Header(),
		oracleHead: head.Header(),
		blocks:     make(map[common.Hash]*types.Block),
		db:         NewOracleBackedIAVLDB(oracle, nil),
	}, nil
}

func (o *OracleBackedPolymeraseChain) CurrentHeader() *types.Header {
	return o.head
}

func (o *OracleBackedPolymeraseChain) GetHeaderByNumber(n uint64) *types.Header {
	if o.head.Number.Uint64() < n {
		return nil
	}
	hash, ok := o.hashByNum[n]
	if ok {
		return o.GetHeaderByHash(hash)
	}
	// Walk back from current head to the requested block number
	h := o.head
	for h.Number.Uint64() > n {
		h = o.GetHeaderByHash(h.ParentHash)
		o.hashByNum[h.Number.Uint64()] = h.Hash()
	}
	o.earliestIndexedBlock = h
	return h
}

func (o *OracleBackedPolymeraseChain) GetTd(hash common.Hash, number uint64) *big.Int {
	// Difficulty is always 0 post-merge and bedrock starts post-merge so total difficulty also always 0
	return common.Big0
}

func (o *OracleBackedPolymeraseChain) CurrentSafeBlock() *types.Header {
	return o.safe
}

func (o *OracleBackedPolymeraseChain) CurrentFinalBlock() *types.Header {
	return o.finalized
}

func (o *OracleBackedPolymeraseChain) GetHeaderByHash(hash common.Hash) *types.Header {
	return o.GetBlockByHash(hash).Header()
}

func (o *OracleBackedPolymeraseChain) GetBlockByHash(hash common.Hash) *types.Block {
	// Check inserted blocks
	block, ok := o.blocks[hash]
	if ok {
		return block
	}
	// Retrieve from the oracle
	return o.oracle.BlockByHash(hash)
}

func (o *OracleBackedPolymeraseChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	var block *types.Block
	if o.oracleHead.Number.Uint64() < number {
		// For blocks above the chain head, only consider newly built blocks
		// Avoids requesting an unknown block from the oracle which would panic.
		block = o.blocks[hash]
	} else {
		block = o.GetBlockByHash(hash)
	}
	if block == nil {
		return nil
	}
	if block.NumberU64() != number {
		return nil
	}
	return block
}

func (o *OracleBackedPolymeraseChain) GetHeader(hash common.Hash, u uint64) *types.Header {
	block := o.GetBlock(hash, u)
	return block.Header()
}

func (o *OracleBackedPolymeraseChain) HasBlockAndState(hash common.Hash, number uint64) bool {
	block := o.GetBlock(hash, number)
	return block != nil
}

func (o *OracleBackedPolymeraseChain) GetCanonicalHash(n uint64) common.Hash {
	header := o.GetHeaderByNumber(n)
	if header == nil {
		return common.Hash{}
	}
	return header.Hash()
}

func (o *OracleBackedPolymeraseChain) Config() *params.ChainConfig {
	return o.chainCfg
}

func (o *OracleBackedPolymeraseChain) Engine() consensus.Engine {
	return o.engine
}

func (o *OracleBackedPolymeraseChain) StateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, state.NewDatabase(rawdb.NewDatabase(o.db)), nil)
}

func (o *OracleBackedPolymeraseChain) InsertBlockWithoutSetHead(block *types.Block) error {
	processor, err := poly_engineapi.NewBlockProcessorFromHeader(o, block.Header())
	if err != nil {
		return err
	}
	for i, tx := range block.Transactions() {
		err = processor.AddTx(tx)
		if err != nil {
			return fmt.Errorf("invalid transaction (%d): %w", i, err)
		}
	}
	expected, err := processor.Assemble()
	if err != nil {
		return fmt.Errorf("invalid block: %w", err)
	}
	if expected.Hash() != block.Hash() {
		return fmt.Errorf("block root mismatch, expected: %v, actual: %v", expected.Hash(), block.Hash())
	}
	err = processor.Commit()
	if err != nil {
		return fmt.Errorf("commit block: %w", err)
	}
	o.blocks[block.Hash()] = block
	return nil
}

func (o *OracleBackedPolymeraseChain) SetCanonical(head *types.Block) (common.Hash, error) {
	oldHead := o.head
	o.head = head.Header()

	// Remove canonical hashes after the new header
	for n := head.NumberU64() + 1; n <= oldHead.Number.Uint64(); n++ {
		delete(o.hashByNum, n)
	}

	// Add new canonical blocks to the block by number cache
	// Since the original head is added to the block number cache and acts as the finalized block,
	// at some point we must reach the existing canonical chain and can stop updating.
	h := o.head
	for {
		newHash := h.Hash()
		prevHash, ok := o.hashByNum[h.Number.Uint64()]
		if ok && prevHash == newHash {
			// Connected with the existing canonical chain so stop updating
			break
		}
		o.hashByNum[h.Number.Uint64()] = newHash
		h = o.GetHeaderByHash(h.ParentHash)
	}
	return head.Hash(), nil
}

func (o *OracleBackedPolymeraseChain) SetFinalized(header *types.Header) {
	o.finalized = header
}

func (o *OracleBackedPolymeraseChain) SetSafe(header *types.Header) {
	o.safe = header
}
