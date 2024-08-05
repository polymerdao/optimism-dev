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
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrMissingL1EthRPC           = errors.New("missing l1 eth rpc url")
	ErrMissingL1Beacon           = errors.New("missing l1 beacon url")
	ErrMissingPlasmaServerRPC    = errors.New("missing op plasma da server url")
	ErrMissingDAChallengeAddress = errors.New("missing da challenge contract address")
	ErrInvalidDACommitmentType   = errors.New("invalid da challenge type")
	ErrMissingDACommitmentKind   = errors.New("missing generic da challenge kind")
	ErrNoActorMode               = errors.New("service must be configured with at least one actor mode")
	ErrMissingL1ClientConfig     = errors.New("missing L1ClientConfig")
	ErrMissingPlasmaConfig       = errors.New("missing PlasmaConfig")
	ErrMissingPlasmaCLIConfig    = errors.New("missing PlasmaCLIConfig")
	ErrMissingRollupConfig       = errors.New("missing RollupConfig")
	ErrMissingBlobSourceConfig   = errors.New("missing BlobSourceConfig")
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
	L1EthRpc        string // L1 RPC Url
	L1Beacon        string // L1 Beacon API Url
	PlasmaServerRpc string // Plasma DA server Url

	Defend    bool
	Challenge bool

	DAChallengeAddress  common.Address
	PollInterval        time.Duration // Polling interval for latest-block subscription when using an HTTP RPC provider
	L1EpochPollInterval time.Duration

	MaxPendingTx uint64 // Maximum number of pending transactions (0 == no limit)
	TxMgrConfig  txmgr.CLIConfig

	MetricsConfig opmetrics.CLIConfig
	PprofConfig   oppprof.CLIConfig

	CommitmentType plasma.CommitmentType
	CommitmentKind CommitmentKind

	L1ClientConfig   *sources.L1ClientConfig
	CLIConfig        *plasma.CLIConfig
	PlasmaConfig     *plasma.Config
	RollupConfig     *rollup.Config
	BlobSourceConfig *node.L1BeaconEndpointConfig
}

func NewConfig(
	defend, challenge bool,
	daChallengeAddress common.Address,
	l1EthRpc string,
	l1BeaconApi string,
	plasmaRPC string,
	commitmentType plasma.CommitmentType,
	commitmentKind CommitmentKind,
	l1ClientConfig *sources.L1ClientConfig,
	l1EpochPollInterval time.Duration,
	plasmaCLIConfig *plasma.CLIConfig,
	plasmaConfig *plasma.Config,
	rollupConfig *rollup.Config,
	blobSourceConfig *node.L1BeaconEndpointConfig,
) Config {
	return Config{
		L1EthRpc:        l1EthRpc,
		L1Beacon:        l1BeaconApi,
		PlasmaServerRpc: plasmaRPC,

		Defend:    defend,
		Challenge: challenge,

		DAChallengeAddress:  daChallengeAddress,
		PollInterval:        DefaultPollInterval,
		L1EpochPollInterval: l1EpochPollInterval,

		MaxPendingTx: DefaultMaxPendingTx,
		TxMgrConfig:  txmgr.NewCLIConfig(l1EthRpc, txmgr.DefaultChallengerFlagValues),

		MetricsConfig: opmetrics.DefaultCLIConfig(),
		PprofConfig:   oppprof.DefaultCLIConfig(),

		CommitmentType: commitmentType,
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
	if c.L1Beacon == "" {
		return ErrMissingL1Beacon
	}
	if c.PlasmaServerRpc == "" {
		return ErrMissingPlasmaServerRPC
	}
	if c.DAChallengeAddress == (common.Address{}) {
		return ErrMissingDAChallengeAddress
	}
	if c.CommitmentType > 1 {
		return ErrInvalidDACommitmentType
	}
	if c.CommitmentType == plasma.GenericCommitmentType {
		if c.CommitmentKind == Undefined {
			return ErrMissingDACommitmentKind
		}
	}
	if c.L1ClientConfig == nil {
		return ErrMissingL1ClientConfig
	}
	if c.L1EpochPollInterval == 0 {
		c.L1EpochPollInterval = DefaultPollInterval
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
