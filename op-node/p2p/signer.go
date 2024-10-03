package p2p

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

var SigningDomainBlocksV1 = [32]byte{}

type Signer interface {
	Sign(ctx context.Context, domain [32]byte, chainID *big.Int, encodedMsg []byte) (sig *[65]byte, err error)
	io.Closer
}

func SigningHash(domain [32]byte, chainID *big.Int, payloadBytes []byte) (common.Hash, error) {
	var msgInput [32 + 32 + 32]byte
	// domain: first 32 bytes
	copy(msgInput[:32], domain[:])
	// chain_id: second 32 bytes
	if chainID.BitLen() > 256 {
		return common.Hash{}, errors.New("chain_id is too large")
	}
	chainID.FillBytes(msgInput[32:64])
	// payload_hash: third 32 bytes, hash of encoded payload
	copy(msgInput[64:], crypto.Keccak256(payloadBytes))

	return crypto.Keccak256Hash(msgInput[:]), nil
}

func BlockSigningHash(cfg *rollup.Config, payloadBytes []byte) (common.Hash, error) {
	return SigningHash(SigningDomainBlocksV1, cfg.L2ChainID, payloadBytes)
}

// LocalSigner is suitable for testing
type LocalSigner struct {
	priv   *ecdsa.PrivateKey
	hasher func(domain [32]byte, chainID *big.Int, payloadBytes []byte) (common.Hash, error)
}

func NewLocalSigner(priv *ecdsa.PrivateKey) *LocalSigner {
	return &LocalSigner{priv: priv, hasher: SigningHash}
}

func (s *LocalSigner) Sign(ctx context.Context, domain [32]byte, chainID *big.Int, encodedMsg []byte) (sig *[65]byte, err error) {
	if s.priv == nil {
		return nil, errors.New("signer is closed")
	}
	signingHash, err := s.hasher(domain, chainID, encodedMsg)
	if err != nil {
		return nil, err
	}
	signature, err := crypto.Sign(signingHash[:], s.priv)
	if err != nil {
		return nil, err
	}
	return (*[65]byte)(signature), nil
}

func (s *LocalSigner) Close() error {
	s.priv = nil
	return nil
}

type RemoteSigner struct {
	endpoint string
	hasher   func(domain [32]byte, chainID *big.Int, payloadBytes []byte) (common.Hash, error)
}

func NewRemoteSigner(endpoint string) *RemoteSigner {
	return &RemoteSigner{endpoint: endpoint, hasher: SigningHash}
}

func (s *RemoteSigner) Sign(ctx context.Context, domain [32]byte, chainID *big.Int, encodedMsg []byte) (sig *[65]byte, err error) {
	signingHash, err := s.hasher(domain, chainID, encodedMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to compute signing hash: %w", err)
	}

	reqBody := struct {
		SigningHash string `json:"signing_hash"`
	}{
		SigningHash: signingHash.Hex(),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody struct {
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	signatureBytes, err := hex.DecodeString(respBody.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	if len(signatureBytes) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signatureBytes))
	}

	return (*[65]byte)(signatureBytes), nil
}

func (s *RemoteSigner) Close() error {
	return nil
}

type PreparedSigner struct {
	Signer
}

func (p *PreparedSigner) SetupSigner(ctx context.Context) (Signer, error) {
	return p.Signer, nil
}

type SignerSetup interface {
	SetupSigner(ctx context.Context) (Signer, error)
}
