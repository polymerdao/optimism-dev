package main

import (
	"context"
	"os"

	"github.com/ethereum-optimism/optimism/op-challenger/version"
	op_dachallenger "github.com/ethereum-optimism/optimism/op-dachallenger"
	"github.com/ethereum-optimism/optimism/op-dachallenger/config"
	daFlags "github.com/ethereum-optimism/optimism/op-dachallenger/flags"
	"github.com/ethereum-optimism/optimism/op-dachallenger/metrics"
	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/flags"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/opio"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	GitCommit = ""
	GitDate   = ""
)

// VersionWithMeta holds the textual version string including the metadata.
var VersionWithMeta = opservice.FormatVersion(version.Version, GitCommit, GitDate, version.Meta)

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Version = VersionWithMeta
	app.Flags = cliapp.ProtectFlags(append(flags.Flags, daFlags.Flags...))
	app.Name = "op-dachallenger"
	app.Usage = "Challenge and resolve DA commitments and challenges"
	app.Description = "Ensures that op-plasma commitments and challenges are correctly challenged and resolved"
	app.Action = cliapp.LifecycleCmd(DAChallenge)
	ctx := opio.WithInterruptBlocker(context.Background())
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func DAChallenge(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	if err := daFlags.CheckRequired(ctx); err != nil {
		log.Crit("missing some required CLI flags", "error", err)
	}
	logCfg := oplog.ReadCLIConfig(ctx)
	l := oplog.NewLogger(oplog.AppOut(ctx), logCfg)
	handler := oplog.NewLogHandler(oplog.AppOut(ctx), logCfg)
	oplog.SetGlobalLogHandler(handler)
	opservice.ValidateEnvVars(daFlags.EnvVarPrefix, daFlags.Flags, l)
	opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, l)
	opservice.WarnOnDeprecatedFlags(ctx, flags.DeprecatedFlags, l)
	m := metrics.NewMetrics()

	ethRPC := ctx.String("l1")
	rpcKindStr := ctx.String("l1-rpckind")
	defend := ctx.Bool("da-defend")
	challenge := ctx.Bool("da-challenge")
	daKind := ctx.Uint("commitment-kind")

	cfg, err := opnode.NewConfig(ctx, l)
	if err != nil {
		log.Crit("unable to create the rollup node config", "error", err)
	}
	cfg.Cancel = closeApp
	var rpcKind sources.RPCProviderKind
	if err := rpcKind.Set(rpcKindStr); err != nil {
		log.Crit("unrecognized rpc kind", "error", err)
	}
	l1ClientCfg := sources.L1ClientDefaultConfig(&cfg.Rollup, true, rpcKind)
	rpCfg, err := cfg.Rollup.GetOPPlasmaConfig()
	if cfg.Plasma.Enabled && err != nil {
		log.Crit("failed to get plasma config", "error", err)
	}
	daCfg := config.NewConfig(defend, challenge, ethRPC, config.CommitmentKind(daKind), l1ClientCfg, &cfg.Plasma,
		&rpCfg, &cfg.Rollup, cfg.Beacon)

	daChallengeService, err := op_dachallenger.Main(ctx.Context, l, daCfg, m)
	if err != nil {
		log.Crit("failed to create a new da challenge service", "error", err)
	}

	if err := daChallengeService.Start(ctx.Context); err != nil {
		log.Crit("failed to start da challenge service", "error", err)
	}

	return daChallengeService, nil
}
