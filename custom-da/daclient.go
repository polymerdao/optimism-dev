package customda

import (
	"context"
	"encoding/hex"
)

// TODO: implement struct
type DAClient struct {
	daFlag string
}

func NewDAClient(daFlag string) *DAClient {
	// TODO: implement constructor
	return &DAClient{
		daFlag,
	}
}

func (c *DAClient) GetInput(ctx context.Context, key []byte) ([]byte, error) {
	var out []byte
	switch key[0] {
	// TODO: implement fetch from custom-da
	case "c"[0]:
		// key[1:] is the blobId
		out = nil
	default:
		// eth fallback
		out = key
	}
	return out, nil
}

func (c *DAClient) SetInput(ctx context.Context, data []byte) ([]byte, error) {
	// TODO: implement blob submission
	blobId, err := "blobId", error(nil)
	var key []byte
	if err == nil {
		// TODO: implement wrapping
		key = []byte(hex.EncodeToString([]byte("c" + blobId)))
	} else {
		// eth fallback
		key = data
	}
	return key, nil
}
