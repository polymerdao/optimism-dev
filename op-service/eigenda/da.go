package eigenda

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"google.golang.org/grpc/credentials"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
)

type IEigenDA interface {
	RetrieveBlob(ctx context.Context, BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error)
	DisperseBlob(ctx context.Context, txData []byte) (*disperser.BlobInfo, error)
}

type EigenDA struct {
	Config

	Log log.Logger
}

func (m *EigenDA) RetrieveBlob(ctx context.Context, BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error) {
	conn, err := m.GetConn()
	if err != nil {
		return nil, err
	}
	daClient := disperser.NewDisperserClient(conn)

	reply, err := daClient.RetrieveBlob(ctx, &disperser.RetrieveBlobRequest{
		BatchHeaderHash: BatchHeaderHash,
		BlobIndex:       BlobIndex,
	})
	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (m *EigenDA) DisperseBlob(ctx context.Context, txData []byte) (*disperser.BlobInfo, error) {
	conn, err := m.GetConn()
	if err != nil {
		m.Log.Error("Unable to connect to EigenDA, aborting")
		return nil, err
	}
	daClient := disperser.NewDisperserClient(conn)

	disperseReq := &disperser.DisperseBlobRequest{
		Data:           txData,
		SecurityParams: m.DisperserSecurityParams,
	}
	disperseRes, err := daClient.DisperseBlob(ctx, disperseReq)

	if err != nil || disperseRes == nil {
		m.Log.Error("Unable to disperse blob to EigenDA, aborting", "err", err)
		return nil, err
	}

	if disperseRes.Result == disperser.BlobStatus_UNKNOWN ||
		disperseRes.Result == disperser.BlobStatus_FAILED {
		m.Log.Error("Unable to disperse blob to EigenDA, aborting", "err", err)
		return nil, fmt.Errorf("reply status is %d", disperseRes.Result)
	}

	base64RequestID := base64.StdEncoding.EncodeToString(disperseRes.RequestId)

	m.Log.Info("Blob disepersed to EigenDA, now waiting for confirmation", "requestID", base64RequestID)

	var statusRes *disperser.BlobStatusReply
	timeoutTime := time.Now().Add(m.StatusQueryTimeout)
	// Wait before first status check
	time.Sleep(m.StatusQueryRetryInterval)
	for time.Now().Before(timeoutTime) {
		statusRes, err = daClient.GetBlobStatus(ctx, &disperser.BlobStatusRequest{
			RequestId: disperseRes.RequestId,
		})
		if err != nil {
			m.Log.Warn("Unable to retrieve blob dispersal status, will retry", "requestID", base64RequestID, "err", err)
		} else if statusRes.Status == disperser.BlobStatus_CONFIRMED {
			// TODO(eigenlayer): As long as fault proofs are disabled, we can move on once a blob is confirmed
			// but not yet finalized, without further logic. Once fault proofs are enabled, we will need to update
			// the proposer to wait until the blob associated with an L2 block has been finalized, i.e. the EigenDA
			// contracts on Ethereum have confirmed the full availability of the blob on EigenDA.
			batchHeaderHashHex := fmt.Sprintf("0x%s", hex.EncodeToString(statusRes.Info.BlobVerificationProof.BatchMetadata.BatchHeaderHash))
			m.Log.Info("Successfully dispersed blob to EigenDA", "requestID", base64RequestID, "batchHeaderHash", batchHeaderHashHex)
			return statusRes.Info, nil
		} else if statusRes.Status == disperser.BlobStatus_UNKNOWN ||
			statusRes.Status == disperser.BlobStatus_FAILED {
			m.Log.Error("EigenDA blob dispersal failed in processing", "requestID", base64RequestID, "err", err)
			return nil, fmt.Errorf("eigenDA blob dispersal failed in processing with reply status %d", statusRes.Status)
		} else {
			m.Log.Warn("Still waiting for confirmation from EigenDA", "requestID", base64RequestID)
		}

		// Wait before first status check
		time.Sleep(m.StatusQueryRetryInterval)
	}

	return nil, fmt.Errorf("timed out getting EigenDA status for dispersed blob key: %s", base64RequestID)
}

func (m *EigenDA) GetConn() (*grpc.ClientConn, error) {
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	options := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(m.RPC, options[0])
	return conn, err
}
