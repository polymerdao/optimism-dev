package config

import (
	"errors"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

var (
	ErrMissingL1EthRPC         = errors.New("missing l1 eth rpc url")
	ErrMissingDACommitmentKind = errors.New("missing generic da challenge kind")
	ErrNoActorMode             = errors.New("service must be configured with at least one actor mode")
	ErrMissingL1ClientConfig   = errors.New("missing L1ClientConfig")
	ErrMissingPlasmaConfig     = errors.New("missing PlasmaConfig")
	ErrMissingPlasmaCLIConfig  = errors.New("missing PlasmaCLIConfig")
	ErrMissingRollupConfig     = errors.New("missing RollupConfig")
	ErrMissingBlobSourceConfig = errors.New("missing BlobSourceConfig")
)

const (
	DefaultPollInterval = time.Second * 12
	DefaultMaxPendingTx = 10
)

type CommitmentKind uint

const (
	Undefined CommitmentKind = iota
	EigenDA
)

// Config is a well typed config that is parsed from the CLI params.
// This also contains config options for auxiliary services.
// It is used to initialize the challenger.
type Config struct {
	L1EthRpc string

	Defend    bool
	Challenge bool

	PollInterval time.Duration // Polling interval for latest-block subscription when using an HTTP RPC provider

	MaxPendingTx uint64 // Maximum number of pending transactions (0 == no limit)
	TxMgrConfig  txmgr.CLIConfig

	MetricsConfig opmetrics.CLIConfig
	PprofConfig   oppprof.CLIConfig

	CommitmentKind CommitmentKind

	L1ClientConfig   *sources.L1ClientConfig
	CLIConfig        *plasma.CLIConfig
	PlasmaConfig     *plasma.Config
	RollupConfig     *rollup.Config
	BlobSourceConfig node.L1BeaconEndpointSetup
}

func NewConfig(
	defend, challenge bool,
	l1EthRpc string,
	commitmentKind CommitmentKind,
	l1ClientConfig *sources.L1ClientConfig,
	plasmaCLIConfig *plasma.CLIConfig,
	plasmaConfig *plasma.Config,
	rollupConfig *rollup.Config,
	blobSourceConfig node.L1BeaconEndpointSetup,
) *Config {
	return &Config{
		L1EthRpc: l1EthRpc,

		Defend:    defend,
		Challenge: challenge,

		PollInterval: DefaultPollInterval,

		MaxPendingTx: DefaultMaxPendingTx,
		TxMgrConfig:  txmgr.NewCLIConfig(l1EthRpc, txmgr.DefaultChallengerFlagValues),

		MetricsConfig: opmetrics.DefaultCLIConfig(),
		PprofConfig:   oppprof.DefaultCLIConfig(),

		CommitmentKind: commitmentKind,

		L1ClientConfig:   l1ClientConfig,
		CLIConfig:        plasmaCLIConfig,
		PlasmaConfig:     plasmaConfig,
		RollupConfig:     rollupConfig,
		BlobSourceConfig: blobSourceConfig,
	}
}

func (c Config) Check() error {
	if !c.Challenge && !c.Defend {
		return ErrNoActorMode
	}
	if c.L1EthRpc == "" {
		return ErrMissingL1EthRPC
	}
	if c.PlasmaConfig.CommitmentType == plasma.GenericCommitmentType {
		if c.CommitmentKind == Undefined {
			return ErrMissingDACommitmentKind
		}
	}
	if c.L1ClientConfig == nil {
		return ErrMissingL1ClientConfig
	}
	if c.PollInterval == 0 {
		c.PollInterval = DefaultPollInterval
	}
	if c.PlasmaConfig == nil {
		return ErrMissingPlasmaConfig
	}
	if c.CLIConfig == nil {
		return ErrMissingPlasmaCLIConfig
	}
	if c.RollupConfig == nil {
		return ErrMissingRollupConfig
	}
	if c.BlobSourceConfig == nil {
		return ErrMissingBlobSourceConfig
	}
	if err := c.TxMgrConfig.Check(); err != nil {
		return err
	}
	if err := c.MetricsConfig.Check(); err != nil {
		return err
	}
	if err := c.PprofConfig.Check(); err != nil {
		return err
	}
	return nil
}
