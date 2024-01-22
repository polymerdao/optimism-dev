package main

import (
	"net"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestShim(t *testing.T) {
	shimPath, err := filepath.Abs("shim")
	require.NoError(t, err)
	require.FileExists(t, shimPath)

	opPolymerPath, err := filepath.Abs("polymer-peptide")
	require.NoError(t, err)
	require.FileExists(t, opPolymerPath)

	config.EthNodeVerbosity = 4

	ec := (&e2e.ExternalRunner{
		Name:    "TestShim",
		BinPath: shimPath,
		L1: eth.BlockID{
			Hash:   common.HexToHash("0x1234"),
			Number: 0,
		},
		L2Time: uint64(time.Now().Unix()),
	}).Run(t)
	t.Cleanup(func() { _ = ec.Close() })

	for _, endpoint := range []string{ec.HTTPEndpoint()} {
		plainURL, err := url.ParseRequestURI(endpoint)
		require.NoError(t, err)
		_, err = net.DialTimeout("tcp", plainURL.Host, time.Second)
		require.NoError(t, err, "could not connect to HTTP port")
	}
	require.NotEqual(t, common.Hash{}, ec.GenesisBlockHash())
}
