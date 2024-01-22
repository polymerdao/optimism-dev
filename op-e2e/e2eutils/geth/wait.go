package geth

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	// errTimeout represents a timeout
	errTimeout = errors.New("timeout")
)

type QueryResult interface {
	*types.Block | *types.Receipt
}

func withTimeout[T QueryResult](timeout time.Duration, cb func(ctx context.Context) (T, error)) (T, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		result, err := cb(ctx)
		if result != nil && err == nil {
			return result, nil
		}
		if err != nil && err.Error() != ethereum.NotFound.Error() {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, errTimeout
		case <-ticker.C:
		}
	}
}

func WaitForL1OriginOnL2(rollupCfg *rollup.Config, l1BlockNum uint64, client *ethclient.Client, timeout time.Duration) (*types.Block, error) {
	return withTimeout[*types.Block](timeout, func(ctx context.Context) (*types.Block, error) {
		block, err := client.BlockByNumber(ctx, nil)
		if err != nil {
			return nil, err
		}
		l1Info, err := derive.L1BlockInfoFromBytes(rollupCfg, block.Time(), block.Transactions()[0].Data())
		if err != nil {
			return nil, err
		}
		if l1Info.Number < l1BlockNum {
			return nil, nil
		}
		return block, nil
	})
}

func WaitForTransaction(hash common.Hash, client *ethclient.Client, timeout time.Duration) (*types.Receipt, error) {
	return withTimeout[*types.Receipt](timeout, func(ctx context.Context) (*types.Receipt, error) {
		return client.TransactionReceipt(ctx, hash)
	})
}

func WaitForBlock(number *big.Int, client *ethclient.Client, timeout time.Duration) (*types.Block, error) {
	return withTimeout[*types.Block](timeout, func(ctx context.Context) (*types.Block, error) {
		return client.BlockByNumber(ctx, number)
	})
}

func WaitForBlockToBeFinalized(number *big.Int, client *ethclient.Client, timeout time.Duration) (*types.Block, error) {
	return waitForBlockTag(number, client, timeout, rpc.FinalizedBlockNumber)
}

func WaitForBlockToBeSafe(number *big.Int, client *ethclient.Client, timeout time.Duration) (*types.Block, error) {
	return waitForBlockTag(number, client, timeout, rpc.SafeBlockNumber)
}

// waitForBlockTag polls for a block number to reach the specified tag & then returns that block at the number.
func waitForBlockTag(number *big.Int, client *ethclient.Client, timeout time.Duration, tag rpc.BlockNumber) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Wait for it to be finalized. Poll every half second.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	tagBigInt := big.NewInt(tag.Int64())

	for {
		select {
		case <-ticker.C:
			block, err := client.BlockByNumber(ctx, tagBigInt)
			if err != nil {
				return nil, err
			}
			if block != nil && block.NumberU64() >= number.Uint64() {
				return client.BlockByNumber(ctx, number)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
