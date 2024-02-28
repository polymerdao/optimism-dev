package driver

import (
	eigenda "github.com/ethereum-optimism/optimism/eigenda"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/log"
)

func SetDAClient(cfg eigenda.CLIConfig, log log.Logger) error {
	client := eigenda.NewDAClient(&cfg, log)
	return derive.SetDAClient(client)
}
