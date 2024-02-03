package sources

import (
	"context"
	crand "crypto/rand"
	"fmt"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
)

type mockRPC struct {
	mock.Mock
}

func (m *mockRPC) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return m.MethodCalled("BatchCallContext", ctx, b).Get(0).([]error)[0]
}

func (m *mockRPC) CallContext(ctx context.Context, result any, method string, args ...any) error {
	return m.MethodCalled("CallContext", ctx, result, method, args).Get(0).([]error)[0]
}

func (m *mockRPC) EthSubscribe(ctx context.Context, channel any, args ...any) (ethereum.Subscription, error) {
	called := m.MethodCalled("EthSubscribe", channel, args)
	return called.Get(0).(*rpc.ClientSubscription), called.Get(1).([]error)[0]
}

func (m *mockRPC) Close() {
	m.MethodCalled("Close")
}

var _ client.RPC = (*mockRPC)(nil)

var testEthClientConfig = &EthClientConfig{
	ReceiptsCacheSize:     10,
	TransactionsCacheSize: 10,
	HeadersCacheSize:      10,
	PayloadsCacheSize:     10,
	MaxRequestsPerBatch:   20,
	MaxConcurrentRequests: 10,
	TrustRPC:              false,
	MustBePostMerge:       false,
	RPCProviderKind:       RPCKindStandard,
}

func randHash() (out common.Hash) {
	_, _ = crand.Read(out[:])
	return out
}

func randHeader() (*types.Header, *rpcHeader) {
	hdr := &types.Header{
		ParentHash:  randHash(),
		UncleHash:   randHash(),
		Coinbase:    common.Address{},
		Root:        randHash(),
		TxHash:      randHash(),
		ReceiptHash: randHash(),
		Bloom:       types.Bloom{},
		Difficulty:  big.NewInt(42),
		Number:      big.NewInt(1234),
		GasLimit:    0,
		GasUsed:     0,
		Time:        123456,
		Extra:       make([]byte, 0),
		MixDigest:   randHash(),
		Nonce:       types.BlockNonce{},
		BaseFee:     big.NewInt(100),
	}
	rhdr := &rpcHeader{
		ParentHash:  hdr.ParentHash,
		UncleHash:   hdr.UncleHash,
		Coinbase:    hdr.Coinbase,
		Root:        hdr.Root,
		TxHash:      hdr.TxHash,
		ReceiptHash: hdr.ReceiptHash,
		Bloom:       eth.Bytes256(hdr.Bloom),
		Difficulty:  *(*hexutil.Big)(hdr.Difficulty),
		Number:      hexutil.Uint64(hdr.Number.Uint64()),
		GasLimit:    hexutil.Uint64(hdr.GasLimit),
		GasUsed:     hexutil.Uint64(hdr.GasUsed),
		Time:        hexutil.Uint64(hdr.Time),
		Extra:       hdr.Extra,
		MixDigest:   hdr.MixDigest,
		Nonce:       hdr.Nonce,
		BaseFee:     (*hexutil.Big)(hdr.BaseFee),
		Hash:        hdr.Hash(),
	}
	return hdr, rhdr
}

func TestEthClient_InfoByHash(t *testing.T) {
	m := new(mockRPC)
	_, rhdr := randHeader()
	expectedInfo, _ := rhdr.Info(true, false)
	ctx := context.Background()
	m.On("CallContext", ctx, new(*rpcHeader),
		"eth_getBlockByHash", []any{rhdr.Hash, false}).Run(func(args mock.Arguments) {
		*args[1].(**rpcHeader) = rhdr
	}).Return([]error{nil})
	s, err := NewEthClient(m, nil, nil, testEthClientConfig)
	require.NoError(t, err)
	info, err := s.InfoByHash(ctx, rhdr.Hash)
	require.NoError(t, err)
	require.Equal(t, info, expectedInfo)
	m.Mock.AssertExpectations(t)
	// Again, without expecting any calls from the mock, the cache will return the block
	info, err = s.InfoByHash(ctx, rhdr.Hash)
	require.NoError(t, err)
	require.Equal(t, info, expectedInfo)
	m.Mock.AssertExpectations(t)
}

func TestEthClient_InfoByNumber(t *testing.T) {
	ctx := context.Background()
	logger := oplog.DefaultCLIConfig()

	log := oplog.NewLogger(os.Stderr, logger)
	l1RPC, err := client.NewRPC(ctx, log, "https://eth-sepolia.g.alchemy.com/v2/zoTGfJGWR7dlEKU2r_ruYFT_AhG_fzZQ", client.WithDialBackoff(5))

	s, err := NewL1Client(l1RPC, log, nil, L1ClientDefaultConfig(&rollup.Config{SeqWindowSize: 10}, false, RPCKindAlchemy))
	require.NoError(t, err)
	info, err := s.InfoByNumber(ctx, uint64(5187053))
	require.NoError(t, err)
	fmt.Print(info)

	_, _, err = s.FetchReceipts(ctx, info.Hash())
	require.NoError(t, err)
}

func TestEthClient_WrongInfoByNumber(t *testing.T) {
	m := new(mockRPC)
	_, rhdr := randHeader()
	rhdr2 := *rhdr
	rhdr2.Number += 1
	n := rhdr.Number
	ctx := context.Background()
	m.On("CallContext", ctx, new(*rpcHeader),
		"eth_getBlockByNumber", []any{n.String(), false}).Run(func(args mock.Arguments) {
		*args[1].(**rpcHeader) = &rhdr2
	}).Return([]error{nil})
	s, err := NewL1Client(m, nil, nil, L1ClientDefaultConfig(&rollup.Config{SeqWindowSize: 10}, true, RPCKindStandard))
	require.NoError(t, err)
	_, err = s.InfoByNumber(ctx, uint64(n))
	require.Error(t, err, "cannot accept the wrong block")
	m.Mock.AssertExpectations(t)
}

func TestEthClient_WrongInfoByHash(t *testing.T) {
	m := new(mockRPC)
	_, rhdr := randHeader()
	rhdr2 := *rhdr
	rhdr2.Root[0] += 1
	rhdr2.Hash = rhdr2.computeBlockHash()
	k := rhdr.Hash
	ctx := context.Background()
	m.On("CallContext", ctx, new(*rpcHeader),
		"eth_getBlockByHash", []any{k, false}).Run(func(args mock.Arguments) {
		*args[1].(**rpcHeader) = &rhdr2
	}).Return([]error{nil})
	s, err := NewL1Client(m, nil, nil, L1ClientDefaultConfig(&rollup.Config{SeqWindowSize: 10}, true, RPCKindStandard))
	require.NoError(t, err)
	_, err = s.InfoByHash(ctx, k)
	require.Error(t, err, "cannot accept the wrong block")
	m.Mock.AssertExpectations(t)
}

type validateReceiptsTest struct {
	desc        string
	trustRPC    bool
	mutReceipts func(types.Receipts) types.Receipts
	expError    string
}

func TestEthClient_validateReceipts(t *testing.T) {
	mutBloom := func(rs types.Receipts) types.Receipts {
		rs[2].Bloom[0] = 1
		return rs
	}
	mutTruncate := func(rs types.Receipts) types.Receipts {
		return rs[:len(rs)-1]
	}

	for _, tt := range []validateReceiptsTest{
		{
			desc: "valid",
		},
		{
			desc:        "invalid-mut-bloom",
			mutReceipts: mutBloom,
			expError:    "invalid receipts: failed to fetch list of receipts: expected receipt root",
		},
		{
			desc:        "invalid-truncated",
			mutReceipts: mutTruncate,
			expError:    "invalid receipts: got 3 receipts but expected 4",
		},
	} {
		t.Run("no-trust-"+tt.desc, func(t *testing.T) {
			testEthClient_validateReceipts(t, tt)
		})
		// trusting the rpc should still lead to failed validation
		tt.trustRPC = true
		t.Run("trust-"+tt.desc, func(t *testing.T) {
			testEthClient_validateReceipts(t, tt)
		})
	}
}

func testEthClient_validateReceipts(t *testing.T, test validateReceiptsTest) {
	require := require.New(t)
	mrpc := new(mockRPC)
	mrp := new(mockReceiptsProvider)
	const numTxs = 4
	block, receipts := randomRpcBlockAndReceipts(rand.New(rand.NewSource(420)), numTxs)
	txHashes := receiptTxHashes(receipts)
	ctx := context.Background()

	if mut := test.mutReceipts; mut != nil {
		receipts = mut(receipts)
	}

	mrpc.On("CallContext", ctx, mock.AnythingOfType("**sources.rpcBlock"),
		"eth_getBlockByHash", []any{block.Hash, true}).
		Run(func(args mock.Arguments) {
			*(args[1].(**rpcBlock)) = block
		}).
		Return([]error{nil}).Once()

	mrp.On("FetchReceipts", ctx, block.BlockID(), txHashes).
		Return(types.Receipts(receipts), error(nil)).Once()

	ethcl := newEthClientWithCaches(nil, numTxs)
	ethcl.client = mrpc
	ethcl.recProvider = mrp
	ethcl.trustRPC = test.trustRPC

	info, recs, err := ethcl.FetchReceipts(ctx, block.Hash)
	if test.expError != "" {
		require.ErrorContains(err, test.expError)
	} else {
		require.NoError(err)
		expInfo, _, err := block.Info(true, false)
		require.NoError(err)
		require.Equal(expInfo, info)
		require.Equal(types.Receipts(receipts), recs)
	}

	mrpc.AssertExpectations(t)
	mrp.AssertExpectations(t)
}

func newEthClientWithCaches(metrics caching.Metrics, cacheSize int) *EthClient {
	return &EthClient{
		transactionsCache: caching.NewLRUCache[common.Hash, types.Transactions](metrics, "txs", cacheSize),
		headersCache:      caching.NewLRUCache[common.Hash, eth.BlockInfo](metrics, "headers", cacheSize),
		payloadsCache:     caching.NewLRUCache[common.Hash, *eth.ExecutionPayloadEnvelope](metrics, "payloads", cacheSize),
	}
}
