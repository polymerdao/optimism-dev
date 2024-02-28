package plasma

import (
	"fmt"
	"math"
	"net/url"
	"time"

	eigenda "github.com/ethereum-optimism/optimism/eigenda"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

const (
	EnabledFlagName                   = "plasma.enabled"
	DaServerAddressFlagName           = "plasma.da-server"
	DaBackendFlagName                 = "plasma.da-backend"
	VerifyOnReadFlagName              = "plasma.verify-on-read"
	PrimaryQuorumIDFlagName           = "plasma.da-primary-quorum-id"
	PrimaryAdversaryThresholdFlagName = "plasma.da-primary-adversary-threshold"
	PrimaryQuorumThresholdFlagName    = "plasma.da-primary-quorum-threshold"
	StatusQueryRetryIntervalFlagName  = "plasma.da-status-query-retry-interval"
	StatusQueryTimeoutFlagName        = "plasma.da-status-query-timeout"
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
		&cli.Uint64Flag{
			Name:    PrimaryAdversaryThresholdFlagName,
			Usage:   "Adversary threshold for the primary quorum of the DA layer",
			EnvVars: plasmaEnv(envPrefix, "DA_PRIMARY_ADVERSARY_THRESHOLD"),
		},
		&cli.Uint64Flag{
			Name:    PrimaryQuorumThresholdFlagName,
			Usage:   "Quorum threshold for the primary quorum of the DA layer",
			EnvVars: plasmaEnv(envPrefix, "DA_PRIMARY_QUORUM_THRESHOLD"),
		},
		&cli.Uint64Flag{
			Name:    PrimaryQuorumIDFlagName,
			Usage:   "Secondary Quorum ID of the DA layer",
			EnvVars: plasmaEnv(envPrefix, "DA_PRIMARY_QUORUM_ID"),
		},
		&cli.DurationFlag{
			Name:    StatusQueryTimeoutFlagName,
			Usage:   "Timeout for aborting an EigenDA blob dispersal if the disperser does not report that the blob has been confirmed dispersed.",
			Value:   1 * time.Minute,
			EnvVars: plasmaEnv(envPrefix, "DA_STATUS_QUERY_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    StatusQueryRetryIntervalFlagName,
			Usage:   "Wait time between retries of EigenDA blob status queries (made while waiting for a blob to be confirmed by)",
			Value:   5 * time.Second,
			EnvVars: plasmaEnv(envPrefix, "DA_STATUS_QUERY_INTERVAL"),
		},
	}
}

type CLIConfig struct {
	Enabled                   bool
	DAServerURL               string
	DABackend                 string
	VerifyOnRead              bool
	PrimaryQuorumID           uint32
	PrimaryAdversaryThreshold uint32
	PrimaryQuorumThreshold    uint32
	StatusQueryRetryInterval  time.Duration
	StatusQueryTimeout        time.Duration
}

func (c CLIConfig) Check() error {
	if !c.Enabled {
		return fmt.Errorf("DA must be enabled; this version requires plasma eigenda")
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
		if c.PrimaryAdversaryThreshold == 0 || c.PrimaryAdversaryThreshold >= 100 {
			return fmt.Errorf("must provide a valid primary DA adversary threshold between (0 and 100)")
		}
		if c.PrimaryQuorumThreshold == 0 || c.PrimaryQuorumThreshold >= 100 {
			return fmt.Errorf("must provide a valid primary DA quorum threshold between (0 and 100)")
		}
		if c.StatusQueryTimeout == 0 {
			return fmt.Errorf("DA status query timeout must be greater than 0")
		}
		if c.StatusQueryRetryInterval == 0 {
			return fmt.Errorf("DA status query retry interval must be greater than 0")
		}
	}
	return nil
}

func (c CLIConfig) NewDAClient(log log.Logger) DAStorage {
	var client DAStorage
	switch c.DABackend {
	case "default":
		client = &DAClient{url: c.DAServerURL, verify: c.VerifyOnRead}
	case "eigenda":
		client = eigenda.NewDAClient(&eigenda.CLIConfig{
			RPC:                       c.DAServerURL,
			PrimaryQuorumID:           c.PrimaryQuorumID,
			PrimaryAdversaryThreshold: c.PrimaryAdversaryThreshold,
			PrimaryQuorumThreshold:    c.PrimaryQuorumThreshold,
			StatusQueryRetryInterval:  c.StatusQueryRetryInterval,
			StatusQueryTimeout:        c.StatusQueryTimeout,
		}, log)
	}
	return client
}

func ReadCLIConfig(c *cli.Context) CLIConfig {
	return CLIConfig{
		Enabled:                   c.Bool(EnabledFlagName),
		DAServerURL:               c.String(DaServerAddressFlagName),
		DABackend:                 c.String(DaBackendFlagName),
		VerifyOnRead:              c.Bool(VerifyOnReadFlagName),
		PrimaryQuorumID:           Uint32(c, PrimaryQuorumIDFlagName),
		PrimaryAdversaryThreshold: Uint32(c, PrimaryAdversaryThresholdFlagName),
		PrimaryQuorumThreshold:    Uint32(c, PrimaryQuorumThresholdFlagName),
		StatusQueryRetryInterval:  c.Duration(StatusQueryRetryIntervalFlagName),
		StatusQueryTimeout:        c.Duration(StatusQueryTimeoutFlagName),
	}
}

// We add this because the urfave/cli library doesn't support uint32 specifically
func Uint32(ctx *cli.Context, flagName string) uint32 {
	daQuorumIDLong := ctx.Uint64(flagName)
	daQuorumID, success := SafeConvertUInt64ToUInt32(daQuorumIDLong)
	if !success {
		panic(fmt.Errorf("%s must be in the uint32 range", flagName))
	}
	return daQuorumID
}

func SafeConvertUInt64ToUInt32(val uint64) (uint32, bool) {
	if val <= math.MaxUint32 {
		return uint32(val), true
	}
	return 0, false
}
