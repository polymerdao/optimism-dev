package op_dachallenger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-challenger/sender"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/keccak256/contracts"
	"github.com/ethereum-optimism/optimism/op-dachallenger/challenge/scheduler"
	"github.com/ethereum-optimism/optimism/op-dachallenger/config"
	"github.com/ethereum-optimism/optimism/op-dachallenger/metrics"
	"github.com/ethereum-optimism/optimism/op-dachallenger/version"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/clock"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/httputil"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

type Service struct {
	logger  log.Logger
	metrics metrics.Metricer
	monitor *gameMonitor
	sched   *scheduler.Scheduler

	txMgr    *txmgr.SimpleTxManager
	txSender *sender.TxSender

	systemClock clock.Clock
	l1Clock     *clock.SimpleClock

	claimer *scheduler.BondClaimScheduler

	daChallengeContract *contracts.DAChallengeContract // how to support variable implementations here?
	rollupClient        *sources.RollupClient

	l1Client   *ethclient.Client
	pollClient client.RPC

	pprofService *oppprof.Service
	metricsSrv   *httputil.HTTPServer

	balanceMetricer io.Closer

	stopped atomic.Bool
}

// NewService creates a new Service.
func NewService(ctx context.Context, logger log.Logger, cfg *config.Config, m metrics.Metricer) (*Service, error) {
	s := &Service{
		systemClock: clock.SystemClock,
		l1Clock:     clock.NewSimpleClock(),
		logger:      logger,
		metrics:     m,
	}

	if err := s.initFromConfig(ctx, cfg); err != nil {
		// upon initialization error we can try to close any of the service components that may have started already.
		return nil, errors.Join(fmt.Errorf("failed to init challenger game service: %w", err), s.Stop(ctx))
	}

	return s, nil
}

func (s *Service) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := s.initTxManager(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init tx manager: %w", err)
	}
	if err := s.initL1Client(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init l1 client: %w", err)
	}
	if err := s.initPollClient(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init poll client: %w", err)
	}
	if err := s.initPProf(&cfg.PprofConfig); err != nil {
		return fmt.Errorf("failed to init profiling: %w", err)
	}
	if err := s.initMetricsServer(&cfg.MetricsConfig); err != nil {
		return fmt.Errorf("failed to init metrics server: %w", err)
	}
	if err := s.initDAChallengeContract(cfg); err != nil {
		return fmt.Errorf("failed to create da challenge contract bindings: %w", err)
	}
	if err := s.initBondClaims(); err != nil {
		return fmt.Errorf("failed to init bond claiming: %w", err)
	}
	if err := s.initScheduler(cfg); err != nil {
		return fmt.Errorf("failed to init scheduler: %w", err)
	}

	s.initMonitor(cfg)

	s.metrics.RecordInfo(version.SimpleWithMeta)
	s.metrics.RecordUp()
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

func (s *Service) initPollClient(ctx context.Context, cfg *config.Config) error {
	pollClient, err := client.NewRPCWithClient(ctx, s.logger, cfg.L1EthRpc, client.NewBaseRPCClient(s.l1Client.Client()), cfg.PollInterval)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	s.pollClient = pollClient
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

func (s *Service) initDAChallengeContract(cfg *config.Config) error {
	daChallengeContract := contracts.NewDAChallengeContract(s.metrics, cfg.DAChallengeAddress,
		batching.NewMultiCaller(s.l1Client.Client(), batching.DefaultBatchSize))
	s.daChallengeContract = daChallengeContract
	return nil
}

func (s *Service) initBondClaims() error {
	claimer := scheduler.NewBondClaimer(s.logger, s.metrics, s.registry.CreateBondContract, s.txSender, s.claimants...)
	s.claimer = scheduler.NewBondClaimScheduler(s.logger, s.metrics, claimer)
	return nil
}

func (s *Service) initScheduler(cfg *config.Config) error {
	disk := newDiskManager(cfg.Datadir)
	s.sched = scheduler.NewScheduler(s.logger, s.metrics, disk, cfg.MaxConcurrency, s.registry.CreatePlayer, cfg.AllowInvalidPrestate)
	return nil
}

func (s *Service) initMonitor(cfg *config.Config) {
	s.monitor = newGameMonitor(s.logger, s.l1Clock, s.factoryContract, s.sched, s.preimages, cfg.GameWindow, s.claimer, cfg.GameAllowlist, s.pollClient)
}

func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("starting scheduler")
	s.sched.Start(ctx)
	s.claimer.Start(ctx)
	s.logger.Info("starting monitoring")
	s.monitor.StartMonitoring()
	s.logger.Info("challenger game service start completed")
	return nil
}

func (s *Service) Stopped() bool {
	return s.stopped.Load()
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("stopping challenger game service")

	var result error
	if s.sched != nil {
		if err := s.sched.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close scheduler: %w", err))
		}
	}
	if s.monitor != nil {
		s.monitor.StopMonitoring()
	}
	if s.claimer != nil {
		if err := s.claimer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close claimer: %w", err))
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

	if s.rollupClient != nil {
		s.rollupClient.Close()
	}
	if s.pollClient != nil {
		s.pollClient.Close()
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
