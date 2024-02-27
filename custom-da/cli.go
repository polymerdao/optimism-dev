package customda

import (
	"fmt"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
)

const (
	DaFlagName = "da.flag"
)

// TODO: implement flags
var (
	defaultDaFlag = "default"
)

func Check(flag string) error {
	//TODO: implement check
	return nil
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    DaFlagName,
			Usage:   "flag",
			Value:   defaultDaFlag,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "DA_FLAG"),
		},
	}
}

type Config struct {
	DaFlag string
}

func (c Config) Check() error {
	if c.DaFlag == "" {
		c.DaFlag = defaultDaFlag
	}

	if err := Check(c.DaFlag); err != nil {
		return fmt.Errorf("invalid da flag: %w", err)
	}

	return nil
}

type CLIConfig struct {
	DaFlag string
}

func (c CLIConfig) Check() error {
	if c.DaFlag == "" {
		c.DaFlag = defaultDaFlag
	}

	if err := Check(c.DaFlag); err != nil {
		return fmt.Errorf("invalid da flag: %w", err)
	}

	return nil
}

func NewCLIConfig() CLIConfig {
	return CLIConfig{
		DaFlag: defaultDaFlag,
	}
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		DaFlag: ctx.String(DaFlagName),
	}
}
