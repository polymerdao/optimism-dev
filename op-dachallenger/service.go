package op_dachallenger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/cliapp"

	"github.com/ethereum-optimism/optimism/op-challenger/sender"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	"github.com/ethereum-optimism/optimism/op-dachallenger/config"
	"github.com/ethereum-optimism/optimism/op-dachallenger/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/version"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/httputil"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

var _ cliapp.Lifecycle = &Service{}

// Service top-level service for op-dachallenger
type Service struct {
	logger  log.Logger
	metrics metrics.Metricer

	registry *challenge.Registry
	actor    *challenge.Actor

	txMgr    *txmgr.SimpleTxManager
	txSender *sender.TxSender

	l1Client            *ethclient.Client
	l1Source            *sources.L1Client
	l1EpochPollInterval time.Duration

	pprofService *oppprof.Service
	metricsSrv   *httputil.HTTPServer

	balanceMetricer io.Closer

	stopped atomic.Bool
}

// NewService creates a new Service.
func NewService(ctx context.Context, logger log.Logger, cfg *config.Config, m metrics.Metricer) (*Service, error) {
	s := &Service{
		logger:              logger,
		metrics:             m,
		l1EpochPollInterval: cfg.PollInterval,
	}

	if err := s.initFromConfig(ctx, cfg); err != nil {
		// upon initialization error we can try to close any of the service components that may have started already.
		return nil, errors.Join(fmt.Errorf("failed to init challenger game service: %w", err), s.Stop(ctx))
	}

	return s, nil
}

func (s *Service) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := s.initPProf(&cfg.PprofConfig); err != nil {
		return fmt.Errorf("failed to init profiling: %w", err)
	}
	if err := s.initTxManager(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init tx manager: %w", err)
	}
	if err := s.initL1Client(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init l1 client: %w", err)
	}
	if err := s.initMetricsServer(&cfg.MetricsConfig); err != nil {
		return fmt.Errorf("failed to init metrics server: %w", err)
	}
	if err := s.initRegistry(cfg); err != nil {
		return fmt.Errorf("failed to create da contract registry: %w", err)
	}
	if err := s.initActor(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init actor: %w", err)
	}

	s.metrics.RecordInfo(version.SimpleWithMeta)
	s.metrics.RecordUp()
	return nil
}

func (s *Service) initPProf(cfg *oppprof.CLIConfig) error {
	s.pprofService = oppprof.New(
		cfg.ListenEnabled,
		cfg.ListenAddr,
		cfg.ListenPort,
		cfg.ProfileType,
		cfg.ProfileDir,
		cfg.ProfileFilename,
	)

	if err := s.pprofService.Start(); err != nil {
		return fmt.Errorf("failed to start pprof service: %w", err)
	}

	return nil
}

func (s *Service) initTxManager(ctx context.Context, cfg *config.Config) error {
	txMgr, err := txmgr.NewSimpleTxManager("challenger", s.logger, s.metrics, cfg.TxMgrConfig)
	if err != nil {
		return fmt.Errorf("failed to create the transaction manager: %w", err)
	}
	s.txMgr = txMgr
	s.txSender = sender.NewTxSender(ctx, s.logger, txMgr, cfg.MaxPendingTx)
	return nil
}

func (s *Service) initL1Client(ctx context.Context, cfg *config.Config) error {
	l1Client, err := dial.DialEthClientWithTimeout(ctx, dial.DefaultDialTimeout, s.logger, cfg.L1EthRpc)
	if err != nil {
		return fmt.Errorf("failed to dial L1: %w", err)
	}
	s.l1Client = l1Client
	return nil
}

func (s *Service) initMetricsServer(cfg *opmetrics.CLIConfig) error {
	if !cfg.Enabled {
		return nil
	}
	s.logger.Debug("starting metrics server", "addr", cfg.ListenAddr, "port", cfg.ListenPort)
	m, ok := s.metrics.(opmetrics.RegistryMetricer)
	if !ok {
		return fmt.Errorf("metrics were enabled, but metricer %T does not expose registry for metrics-server", s.metrics)
	}
	metricsSrv, err := opmetrics.StartServer(m.Registry(), cfg.ListenAddr, cfg.ListenPort)
	if err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}
	s.logger.Info("started metrics server", "addr", metricsSrv.Addr())
	s.metricsSrv = metricsSrv
	s.balanceMetricer = s.metrics.StartBalanceMetrics(s.logger, s.l1Client, s.txSender.From())
	return nil
}

func (s *Service) initRegistry(cfg *config.Config) error {
	creator, err := challenge.NewContractCreator(cfg.PlasmaConfig.CommitmentType, cfg.CommitmentKind)
	if err != nil {
		return err
	}
	s.registry = &challenge.Registry{}
	s.registry.RegisterBondUnlockContract(cfg.PlasmaConfig.CommitmentType, cfg.CommitmentKind,
		creator.UnlockContractCreator)
	s.registry.RegisterChallengeResolverContract(cfg.PlasmaConfig.CommitmentType, cfg.CommitmentKind,
		creator.ResolveContractCreator)
	s.registry.RegisterBondWithdrawContract(cfg.PlasmaConfig.CommitmentType, cfg.CommitmentKind,
		creator.WithdrawContractCreator)
	s.registry.RegisterCommitmentChallengeContract(cfg.PlasmaConfig.CommitmentType, cfg.CommitmentKind,
		creator.ChallengeContractCreator)
	return nil
}

func (s *Service) initActor(ctx context.Context, cfg *config.Config) error {
	pollClient, err := client.NewRPCWithClient(ctx, s.logger, cfg.L1EthRpc, client.NewBaseRPCClient(s.l1Client.Client()), cfg.PollInterval)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	l1Source, err := sources.NewL1Client(pollClient, s.logger, s.metrics, cfg.L1ClientConfig)
	if err != nil {
		return err
	}

	damgr := plasma.NewPlasmaDA(s.logger, *cfg.CLIConfig, *cfg.PlasmaConfig, &plasma.NoopMetrics{}) // TODO: add DA metrics

	beaconClient, fallbacks, err := cfg.BlobSourceConfig.Setup(ctx, s.logger)
	if err != nil {
		return fmt.Errorf("failed to setup L1 Beacon API client: %w", err)
	}
	beaconCfg := sources.L1BeaconClientConfig{
		FetchAllSidecars: cfg.BlobSourceConfig.ShouldFetchAllSidecars(),
	}
	beacon := sources.NewL1BeaconClient(beaconClient, beaconCfg, fallbacks...)

	var defender *scheduler.ResolveScheduler
	var challenger *scheduler.ChallengeScheduler
	if cfg.Defend {
		contract, err := s.registry.
			CreateResolveContract[cfg.PlasmaConfig.CommitmentType][cfg.CommitmentKind](context.Background(), s.metrics,
			cfg.PlasmaConfig.DAChallengeContractAddress, batching.NewMultiCaller(s.l1Client.Client(),
				batching.DefaultBatchSize))
		if err != nil {
			return err
		}
		res := scheduler.NewResolver(s.logger, contract, s.txSender)
		defender = scheduler.NewResolveScheduler(s.logger, s.metrics, res)
	}
	if cfg.Challenge {
		contract, err := s.registry.
			CreateChallengeContract[cfg.PlasmaConfig.CommitmentType][cfg.CommitmentKind](context.Background(), s.metrics,
			cfg.PlasmaConfig.DAChallengeContractAddress, batching.NewMultiCaller(s.l1Client.Client(),
				batching.DefaultBatchSize))
		if err != nil {
			return err
		}
		chal := scheduler.NewChallenger(s.logger, contract, s.txSender)
		challenger = scheduler.NewChallengeScheduler(s.logger, s.metrics, chal)
	}

	actor, err := challenge.NewActor(s.logger, s.metrics, defender, challenger,
		damgr, l1Source, beacon, cfg.RollupConfig)
	if err != nil {
		return err
	}
	s.actor = actor
	s.l1Source = l1Source
	return nil
}

// Start satisfies cliapp.Lifecycle
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("starting da challenge service")
	s.actor.Start(ctx)
	finalHeadSub := eth.PollBlockChanges(s.logger, s.l1Source, s.actor.OnNewL1Finalized, eth.Finalized,
		s.l1EpochPollInterval, time.Second*10)
	go func() {
		for {
			select {
			case err := <-finalHeadSub.Err():
				s.logger.Error("finalHeadSub error:", err)
			default:
				if s.stopped.Load() {
					return
				}
			}
		}
	}()
	s.logger.Info("challenger game service start completed")
	return nil
}

// Stopped satisfies cliapp.Lifecycle
func (s *Service) Stopped() bool {
	return s.stopped.Load()
}

// Stop satisfied cliapp.Lifecycle
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("stopping da challenger service")

	var result error
	if s.actor != nil {
		if err := s.actor.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close coordinator: %w", err))
		}
	}
	if s.pprofService != nil {
		if err := s.pprofService.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close pprof server: %w", err))
		}
	}
	if s.balanceMetricer != nil {
		if err := s.balanceMetricer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close balance metricer: %w", err))
		}
	}
	if s.txMgr != nil {
		s.txMgr.Close()
	}
	if s.l1Client != nil {
		s.l1Client.Close()
	}
	if s.metricsSrv != nil {
		if err := s.metricsSrv.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close metrics server: %w", err))
		}
	}
	s.stopped.Store(true)

	s.logger.Info("stopped challenger game service", "err", result)
	return result
}
