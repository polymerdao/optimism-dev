package plasma

import (
	"fmt"
	"net/url"

	customda "github.com/ethereum-optimism/optimism/custom-da"
	"github.com/urfave/cli/v2"
)

const (
	EnabledFlagName         = "plasma.enabled"
	DaServerAddressFlagName = "plasma.da-server"
	DaBackendFlagName       = "plasma.da-backend"
	VerifyOnReadFlagName    = "plasma.verify-on-read"
)

func plasmaEnv(envprefix, v string) []string {
	return []string{envprefix + "_PLASMA_" + v}
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    EnabledFlagName,
			Usage:   "Enable plasma mode",
			Value:   false,
			EnvVars: plasmaEnv(envPrefix, "ENABLED"),
		},
		&cli.StringFlag{
			Name:    DaServerAddressFlagName,
			Usage:   "HTTP address of a DA Server",
			EnvVars: plasmaEnv(envPrefix, "DA_SERVER"),
		},
		&cli.StringFlag{
			Name:    DaBackendFlagName,
			Usage:   "Plamsa mode backend ('eigenda')",
			EnvVars: plasmaEnv(envPrefix, "DA_BACKEND"),
		},
		&cli.BoolFlag{
			Name:    VerifyOnReadFlagName,
			Usage:   "Verify input data matches the commitments from the DA storage service",
			Value:   true,
			EnvVars: plasmaEnv(envPrefix, "VERIFY_ON_READ"),
		},
	}
}

type CLIConfig struct {
	Enabled      bool
	DAServerURL  string
	DABackend    string
	VerifyOnRead bool
}

func (c CLIConfig) Check() error {
	if !c.Enabled {
		return fmt.Errorf("DA must be enabled; this version requires plasma customda")
	}
	if c.Enabled {
		if c.DAServerURL == "" {
			return fmt.Errorf("DA server URL is required when plasma da is enabled")
		}
		if _, err := url.Parse(c.DAServerURL); err != nil {
			return fmt.Errorf("DA server URL is invalid: %w", err)
		}
		if c.DABackend != "eigenda" {
			return fmt.Errorf("DA backend unsupported; this version requires plasma eigenda")
		}
	}
	return nil
}

func (c CLIConfig) NewDAClient() DAStorage {
	var client DAStorage
	switch c.DABackend {
	case "default":
		client = &DAClient{url: c.DAServerURL, verify: c.VerifyOnRead}
	case "customda":
		client = customda.NewDAClient(c.DAServerURL /*c.VerifyOnRead*/)
	}
	return client
}

func ReadCLIConfig(c *cli.Context) CLIConfig {
	return CLIConfig{
		Enabled:      c.Bool(EnabledFlagName),
		DAServerURL:  c.String(DaServerAddressFlagName),
		DABackend:    c.String(DaBackendFlagName),
		VerifyOnRead: c.Bool(VerifyOnReadFlagName),
	}
}
