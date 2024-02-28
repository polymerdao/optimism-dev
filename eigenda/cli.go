package eigenda

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
)

const (
	RPCFlagName                       = "da-rpc"
	PrimaryQuorumIDFlagName           = "da-primary-quorum-id"
	PrimaryAdversaryThresholdFlagName = "da-primary-adversary-threshold"
	PrimaryQuorumThresholdFlagName    = "da-primary-quorum-threshold"
	StatusQueryRetryIntervalFlagName  = "da-status-query-retry-interval"
	StatusQueryTimeoutFlagName        = "da-status-query-timeout"
)

type Config struct {
	// TODO(eigenlayer): Update quorum ID command-line parameters to support passing
	// and arbitrary number of quorum IDs.

	// DaRpc is the HTTP provider URL for the Data Availability node.
	RPC string

	// Quorum IDs and SecurityParams to use when dispersing and retrieving blobs
	DisperserSecurityParams []*disperser.SecurityParams

	// The total amount of time that the batcher will spend waiting for EigenDA to confirm a blob
	StatusQueryTimeout time.Duration

	// The amount of time to wait between status queries of a newly dispersed blob
	StatusQueryRetryInterval time.Duration
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

type CLIConfig struct {
	RPC                       string
	PrimaryQuorumID           uint32
	PrimaryAdversaryThreshold uint32
	PrimaryQuorumThreshold    uint32
	StatusQueryRetryInterval  time.Duration
	StatusQueryTimeout        time.Duration
}

func (c CLIConfig) Check() error {
	if c.RPC == "" {
		return errors.New("must provide a DA RPC url")
	}
	if c.PrimaryAdversaryThreshold == 0 || c.PrimaryAdversaryThreshold >= 100 {
		return errors.New("must provide a valid primary DA adversary threshold between (0 and 100)")
	}
	if c.PrimaryQuorumThreshold == 0 || c.PrimaryQuorumThreshold >= 100 {
		return errors.New("must provide a valid primary DA quorum threshold between (0 and 100)")
	}
	if c.StatusQueryTimeout == 0 {
		return errors.New("DA status query timeout must be greater than 0")
	}
	if c.StatusQueryRetryInterval == 0 {
		return errors.New("DA status query retry interval must be greater than 0")
	}
	return nil
}

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
