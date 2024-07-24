package config

import (
	"errors"
	"github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"runtime"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/sources"

	plasma "github.com/ethereum-optimism/optimism/op-plasma"

	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrMaxConcurrencyZero        = errors.New("max concurrency must not be 0")
	ErrMissingL1EthRPC           = errors.New("missing l1 eth rpc url")
	ErrMissingL1Beacon           = errors.New("missing l1 beacon url")
	ErrMissingPlasmaServerRPC    = errors.New("missing op plasma da server url")
	ErrMissingBatchInboxAddress  = errors.New("missing batch inbox address")
	ErrMissingDAChallengeAddress = errors.New("missing da challenge contract address")
	ErrInvalidDACommitmentType   = errors.New("invalid da challenge type")
	ErrMissingDACommitmentKind   = errors.New("missing generic da challenge kind")
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
	L1EthRpc            string // L1 RPC Url
	L1Beacon            string // L1 Beacon API Url
	PlasmaServerRpc     string // Plasma DA server Url

	BatchInboxAddress   common.Address // Address of the dispute game factory
	DAChallengeAddress  common.Address
	BatcherAddresses    []common.Address // Allowed batch submitters, ignore all other
	MaxConcurrency      uint             // Maximum number of threads to use when progressing games
	PollInterval        time.Duration    // Polling interval for latest-block subscription when using an HTTP RPC provider
	L1ClientConfig      *sources.L1ClientConfig
	L1EpochPollInterval time.Duration

	MaxPendingTx uint64 // Maximum number of pending transactions (0 == no limit)

	TxMgrConfig   txmgr.CLIConfig
	MetricsConfig opmetrics.CLIConfig
	PprofConfig   oppprof.CLIConfig

	CommitmentType plasma.CommitmentType
	CommitmentKind CommitmentKind

	CLIConfig    *plasma.CLIConfig
	PlasmaConfig *plasma.Config
	RollupConfig *rollup.Config
	BlobSourceConfig *node.L1BeaconEndpointConfig
}

func NewConfig(
	batchInboxAddress common.Address,
	daChallengeAddress common.Address,
	l1EthRpc string,
	l1BeaconApi string,
	plasmaRPC string,
	batcherAddrs []common.Address,
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
		L1EthRpc:            l1EthRpc,
		L1Beacon:            l1BeaconApi,
		PlasmaServerRpc:     plasmaRPC,
		BatchInboxAddress:   batchInboxAddress,
		BatcherAddresses:    batcherAddrs,
		DAChallengeAddress:  daChallengeAddress,
		MaxConcurrency:      uint(runtime.NumCPU()),
		PollInterval:        DefaultPollInterval,
		L1ClientConfig:      l1ClientConfig,
		L1EpochPollInterval: l1EpochPollInterval,

		MaxPendingTx: DefaultMaxPendingTx,

		TxMgrConfig:   txmgr.NewCLIConfig(l1EthRpc, txmgr.DefaultChallengerFlagValues),
		MetricsConfig: opmetrics.DefaultCLIConfig(),
		PprofConfig:   oppprof.DefaultCLIConfig(),

		CommitmentType: commitmentType,
		CommitmentKind: commitmentKind,

		CLIConfig:    plasmaCLIConfig,
		PlasmaConfig: plasmaConfig,
		RollupConfig: rollupConfig,
		BlobSourceConfig: blobSourceConfig,
	}
}

func (c Config) Check() error {
	if c.L1EthRpc == "" {
		return ErrMissingL1EthRPC
	}
	if c.L1Beacon == "" {
		return ErrMissingL1Beacon
	}
	if c.PlasmaServerRpc == "" {
		return ErrMissingPlasmaServerRPC
	}
	if c.BatchInboxAddress == (common.Address{}) {
		return ErrMissingBatchInboxAddress
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
		c.L1EpochPollInterval = DefaultL1EpochPollInterval
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
	if c.MaxConcurrency == 0 {
		return ErrMaxConcurrencyZero
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
