package eigenda

import (
	"errors"
	"time"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/urfave/cli/v2"
)

const (
	RPCFlagName                       = "da-rpc"
	PrimaryQuorumIDFlagName           = "da-primary-quorum-id"
	PrimaryAdversaryThresholdFlagName = "da-primary-adversary-threshold"
	PrimaryQuorumThresholdFlagName    = "da-primary-quorum-threshold"
	StatusQueryRetryIntervalFlagName  = "da-status-query-retry-interval"
	StatusQueryTimeoutFlagName        = "da-status-query-timeout"
)

type CLIConfig struct {
	RPC                       string
	PrimaryQuorumID           uint32
	PrimaryAdversaryThreshold uint32
	PrimaryQuorumThreshold    uint32
	StatusQueryRetryInterval  time.Duration
	StatusQueryTimeout        time.Duration
}

// NewConfig parses the Config from the provided flags or environment variables.
func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		/* Required Flags */
		RPC:                       ctx.String(RPCFlagName),
		PrimaryQuorumID:           Uint32(ctx, PrimaryQuorumIDFlagName),
		PrimaryAdversaryThreshold: Uint32(ctx, PrimaryAdversaryThresholdFlagName),
		PrimaryQuorumThreshold:    Uint32(ctx, PrimaryQuorumThresholdFlagName),
		StatusQueryRetryInterval:  ctx.Duration(StatusQueryRetryIntervalFlagName),
		StatusQueryTimeout:        ctx.Duration(StatusQueryTimeoutFlagName),
	}
}

func (m CLIConfig) Check() error {
	if m.RPC == "" {
		return errors.New("must provide a DA RPC url")
	}
	if m.PrimaryAdversaryThreshold == 0 || m.PrimaryAdversaryThreshold >= 100 {
		return errors.New("must provide a valid primary DA adversary threshold between (0 and 100)")
	}
	if m.PrimaryQuorumThreshold == 0 || m.PrimaryQuorumThreshold >= 100 {
		return errors.New("must provide a valid primary DA quorum threshold between (0 and 100)")
	}
	if m.StatusQueryTimeout == 0 {
		return errors.New("DA status query timeout must be greater than 0")
	}
	if m.StatusQueryRetryInterval == 0 {
		return errors.New("DA status query retry interval must be greater than 0")
	}
	return nil
}

func CLIFlags(envPrefix string) []cli.Flag {
	prefixEnvVars := func(name string) []string {
		return opservice.PrefixEnvVar(envPrefix, name)
	}
	return []cli.Flag{
		&cli.StringFlag{
			Name:    RPCFlagName,
			Usage:   "RPC endpoint of the EigenDA disperser",
			EnvVars: prefixEnvVars("DA_RPC"),
		},
		&cli.Uint64Flag{
			Name:    PrimaryAdversaryThresholdFlagName,
			Usage:   "Adversary threshold for the primary quorum of the DA layer",
			EnvVars: prefixEnvVars("DA_PRIMARY_ADVERSARY_THRESHOLD"),
		},
		&cli.Uint64Flag{
			Name:    PrimaryQuorumThresholdFlagName,
			Usage:   "Quorum threshold for the primary quorum of the DA layer",
			EnvVars: prefixEnvVars("DA_PRIMARY_QUORUM_THRESHOLD"),
		},
		&cli.Uint64Flag{
			Name:    PrimaryQuorumIDFlagName,
			Usage:   "Secondary Quorum ID of the DA layer",
			EnvVars: prefixEnvVars("DA_PRIMARY_QUORUM_ID"),
		},
		&cli.DurationFlag{
			Name:    StatusQueryTimeoutFlagName,
			Usage:   "Timeout for aborting an EigenDA blob dispersal if the disperser does not report that the blob has been confirmed dispersed.",
			Value:   1 * time.Minute,
			EnvVars: prefixEnvVars("DA_STATUS_QUERY_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    StatusQueryRetryIntervalFlagName,
			Usage:   "Wait time between retries of EigenDA blob status queries (made while waiting for a blob to be confirmed by)",
			Value:   5 * time.Second,
			EnvVars: prefixEnvVars("DA_STATUS_QUERY_INTERVAL"),
		},
	}
}
