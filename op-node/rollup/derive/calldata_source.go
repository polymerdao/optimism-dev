package derive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/protobuf/proto"
)

// CalldataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type CalldataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	ref     eth.L1BlockRef
	dsCfg   DataSourceConfig
	fetcher L1TransactionFetcher
	log     log.Logger

	batcherAddr common.Address
	daClient    eigenda.IEigenDA
}

// NewCalldataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewCalldataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, fetcher L1TransactionFetcher, ref eth.L1BlockRef, batcherAddr common.Address, daClient eigenda.IEigenDA) DataIter {
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, ref.Hash)
	if err != nil {
		return &CalldataSource{
			open:        false,
			ref:         ref,
			dsCfg:       dsCfg,
			fetcher:     fetcher,
			log:         log,
			batcherAddr: batcherAddr,
			daClient:    daClient,
		}
	}
	return &CalldataSource{
		open: true,
		data: DataFromEVMTransactions(dsCfg, batcherAddr, txs, log.New("origin", ref), daClient),
	}
}

// Next returns the next piece of data if it has it. If the constructor failed, this
// will attempt to reinitialize itself. If it cannot find the block it returns a ResetError
// otherwise it returns a temporary error if fetching the block returns an error.
func (ds *CalldataSource) Next(ctx context.Context) (eth.Data, error) {
	if !ds.open {
		if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.ref.Hash); err == nil {
			ds.open = true
			ds.data = DataFromEVMTransactions(ds.dsCfg, ds.batcherAddr, txs, ds.log, ds.daClient)
		} else if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open calldata source: %w", err))
		} else {
			return nil, NewTemporaryError(fmt.Errorf("failed to open calldata source: %w", err))
		}
	}
	if len(ds.data) == 0 {
		return nil, io.EOF
	} else {
		data := ds.data[0]
		ds.data = ds.data[1:]
		return data, nil
	}
}

// DataFromEVMTransactions filters all of the transactions and returns the calldata from transactions
// that are sent to the batch inbox address from the batch sender address.
// This will return an empty array if no valid transactions are found.
func DataFromEVMTransactions(dsCfg DataSourceConfig, batcherAddr common.Address, txs types.Transactions, log log.Logger, daClient eigenda.IEigenDA) []eth.Data {
	out := []eth.Data{}
	for j, tx := range txs {
		if isValidBatchTx(tx, dsCfg.l1Signer, dsCfg.batchInboxAddress, batcherAddr) {
			calldataFrame := &op_service.CalldataFrame{}
			err := proto.Unmarshal(tx.Data(), calldataFrame)
			if err != nil {
				log.Warn("unable to decode calldata frame", "index", j, "err", err)
				return nil
			}

			switch calldataFrame.Value.(type) {
			case *op_service.CalldataFrame_FrameRef:
				frameRef := calldataFrame.GetFrameRef()
				if len(frameRef.QuorumIds) == 0 {
					log.Warn("decoded frame ref contains no quorum IDs", "index", j, "err", err)
					return nil
				}

				log.Info("requesting data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
				data, err := daClient.RetrieveBlob(context.Background(), frameRef.BatchHeaderHash, frameRef.BlobIndex)
				if err != nil {
					retrieveReqJSON, _ := json.Marshal(struct {
						BatchHeaderHash      string
						BlobIndex            uint32
						ReferenceBlockNumber uint32
						QuorumId             uint32
					}{
						BatchHeaderHash:      base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash),
						BlobIndex:            frameRef.BlobIndex,
						ReferenceBlockNumber: frameRef.ReferenceBlockNumber,
						QuorumId:             frameRef.QuorumIds[0],
					})
					log.Warn("could not retrieve data from EigenDA", "request", string(retrieveReqJSON), "err", err)
					return nil
				}
				log.Info("Successfully retrieved data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
				data = data[:frameRef.BlobLength]
				out = append(out, data)
			case *op_service.CalldataFrame_Frame:
				log.Info("Successfully read data from calldata (not EigenDA)")
				frame := calldataFrame.GetFrame()
				out = append(out, frame)
			}

			out = append(out, tx.Data())
		}
	}
	return out
}
