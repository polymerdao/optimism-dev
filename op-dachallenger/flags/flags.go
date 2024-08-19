package flags

import (
	"fmt"

	"github.com/urfave/cli/v2"

	opflags "github.com/ethereum-optimism/optimism/op-service/flags"
)

const EnvVarPrefix = "DA_CHALLENGE"

const (
	DACategory = "1. DA"
)

func prefixEnvVars(names ...string) []string {
	envs := make([]string, 0, len(names))
	for _, name := range names {
		envs = append(envs, EnvVarPrefix+"_"+name)
	}
	return envs
}

var (
	DADefend = &cli.BoolFlag{
		Name:     "da-defend",
		Usage:    "Turns on DA challenge resolving",
		Value:    true,
		EnvVars:  prefixEnvVars("DA_DEFEND"),
		Category: DACategory,
	}
	DAChallenge = &cli.BoolFlag{
		Name:     "da-challenge",
		Usage:    "Turns on DA commitment challenging",
		EnvVars:  prefixEnvVars("DA_CHALLENGE"),
		Value:    true,
		Category: DACategory,
	}
	CommitmentKind = &cli.UintFlag{
		Name:     "commitment-kind",
		Usage:    "Kind of DA commitment",
		Value:    0,
		EnvVars:  prefixEnvVars("COMMITMENT_KIND"),
		Category: DACategory,
	}
)

var requiredFlags = []cli.Flag{
	DADefend,
	DAChallenge,
	CommitmentKind,
}

var optionalFlags = []cli.Flag{}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return opflags.CheckRequiredXor(ctx)
}
