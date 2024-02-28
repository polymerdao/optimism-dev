package eigenda

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/protobuf/proto"
)

// TODO: implement struct
type DAClient struct {
	eigenDaClient IEigenDA
}

func NewDAClient(daCfg *CLIConfig, log log.Logger) *DAClient {
	disperserSecurityParams := []*disperser.SecurityParams{}
	disperserSecurityParams = append(disperserSecurityParams, &disperser.SecurityParams{
		QuorumId:           daCfg.PrimaryQuorumID,
		AdversaryThreshold: daCfg.PrimaryAdversaryThreshold,
		QuorumThreshold:    daCfg.PrimaryQuorumThreshold,
	})

	return &DAClient{
		eigenDaClient: &EigenDA{
			Config: Config{
				RPC:                      daCfg.RPC,
				DisperserSecurityParams:  disperserSecurityParams,
				StatusQueryTimeout:       daCfg.StatusQueryTimeout,
				StatusQueryRetryInterval: daCfg.StatusQueryRetryInterval,
			},
			Log: log,
		},
	}
}

func (c *DAClient) GetInput(ctx context.Context, key []byte) ([]byte, error) {
	var out []byte

	calldataFrame := &op_service.CalldataFrame{}
	err := proto.Unmarshal(key, calldataFrame)
	if err != nil {
		log.Warn("unable to decode calldata frame", "err", err)
		return nil, err
	}

	switch calldataFrame.Value.(type) {
	case *op_service.CalldataFrame_FrameRef:
		frameRef := calldataFrame.GetFrameRef()
		if len(frameRef.QuorumIds) == 0 {
			log.Warn("decoded frame ref contains no quorum IDs", "err", err)
			return nil, err
		}

		log.Info("requesting data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
		data, err := c.eigenDaClient.RetrieveBlob(context.Background(), frameRef.BatchHeaderHash, frameRef.BlobIndex)
		if err != nil {
			retrieveReqJSON, _ := json.Marshal(struct {
				BatchHeaderHash string
				BlobIndex       uint32
			}{
				BatchHeaderHash: base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash),
				BlobIndex:       frameRef.BlobIndex,
			})
			log.Warn("could not retrieve data from EigenDA", "request", string(retrieveReqJSON), "err", err)
			return nil, err
		}
		log.Info("Successfully retrieved data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
		out = data[:frameRef.BlobLength]
	case *op_service.CalldataFrame_Frame:
		log.Info("Successfully read data from calldata (not EigenDA)")
		frame := calldataFrame.GetFrame()
		out = append(out, frame...)
	}
	return out, nil
}

func (c *DAClient) SetInput(ctx context.Context, data []byte) ([]byte, error) {
	blobInfo, err := c.eigenDaClient.DisperseBlob(context.Background(), data)
	var key []byte
	if err == nil {
		quorumIDs := make([]uint32, len(blobInfo.BlobHeader.BlobQuorumParams))
		for i := range quorumIDs {
			quorumIDs[i] = blobInfo.BlobHeader.BlobQuorumParams[i].QuorumNumber
		}
		calldataFrame := &op_service.CalldataFrame{
			Value: &op_service.CalldataFrame_FrameRef{
				FrameRef: &op_service.FrameRef{
					BatchHeaderHash:      blobInfo.BlobVerificationProof.BatchMetadata.BatchHeaderHash,
					BlobIndex:            blobInfo.BlobVerificationProof.BlobIndex,
					ReferenceBlockNumber: blobInfo.BlobVerificationProof.BatchMetadata.BatchHeader.ReferenceBlockNumber,
					QuorumIds:            quorumIDs,
					BlobLength:           uint32(len(data)),
				},
			},
		}
		key, err = proto.Marshal(calldataFrame)
		if err != nil {
			return nil, err
		}
	} else {
		// eth fallback
		key = data
	}
	return key, nil
}
