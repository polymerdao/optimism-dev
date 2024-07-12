package op_dachallenger

import (
	"context"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-dachallenger/config"
	"github.com/ethereum-optimism/optimism/op-dachallenger/metrics"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
)

// Main is the programmatic entry-point for running op-challenger with a given configuration.
func Main(ctx context.Context, logger log.Logger, cfg *config.Config, m metrics.Metricer) (cliapp.Lifecycle, error) {
	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return NewService(ctx, logger, cfg, m)
}
