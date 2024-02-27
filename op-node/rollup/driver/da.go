package driver

import (
	customda "github.com/ethereum-optimism/optimism/custom-da"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
)

func SetDAClient(cfg customda.Config) error {
	client := customda.NewDAClient(cfg.DaFlag)
	return derive.SetDAClient(client)
}
