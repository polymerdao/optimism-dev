package metrics

import (
	"io"

	contractMetrics "github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/httputil"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	txmetrics "github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"
)

const Namespace = "op_dachallenger"

type Metricer interface {
	RecordUp()
	RecordInfo(version string)

	StartBalanceMetrics(l log.Logger, client *ethclient.Client, account common.Address) io.Closer

	// Record Tx metrics
	txmetrics.TxMetricer

	// Record cache metrics
	caching.Metrics

	// Record contract metrics
	contractMetrics.ContractMetricer

	RecordLastActedL1Block(n uint64)
	RecordActedL1Ref(ref eth.L1BlockRef)

	RecordDAChallenge()
	RecordDAChallengeFailed()
	RecordDAResolve()
	RecordDAResolveFailed()
	RecordBondUnlock()
	RecordBondUnlockFailed()
	RecordWithdraw()
	RecordWithdrawFailed()
}

// Metrics implementation must implement RegistryMetricer to allow the metrics server to work.
var _ opmetrics.RegistryMetricer = (*Metrics)(nil)

type Metrics struct {
	ns       string
	registry *prometheus.Registry
	factory  opmetrics.Factory

	txmetrics.TxMetrics
	*opmetrics.CacheMetrics
	*contractMetrics.ContractMetrics

	highestLastActedL1Block prometheus.Gauge
	recordedL1Ref           prometheus.GaugeVec

	up   prometheus.Counter
	info prometheus.GaugeVec

	challenges        prometheus.Counter
	failedChallenges  prometheus.Counter
	resolutions       prometheus.Counter
	failedResolutions prometheus.Counter
	withdraws         prometheus.Counter
	failedWithdraws   prometheus.Counter
	unlocks           prometheus.Counter
	failedUnlocks     prometheus.Counter
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

var _ Metricer = (*Metrics)(nil)

func NewMetrics() *Metrics {
	registry := opmetrics.NewRegistry()
	factory := opmetrics.With(registry)

	return &Metrics{
		ns:       Namespace,
		registry: registry,
		factory:  factory,

		TxMetrics: txmetrics.MakeTxMetrics(Namespace, factory),

		CacheMetrics: opmetrics.NewCacheMetrics(factory, Namespace, "provider_cache", "Provider cache"),

		ContractMetrics: contractMetrics.MakeContractMetrics(Namespace, factory),

		info: *factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "info",
			Help:      "Pseudo-metric tracking version and config info",
		}, []string{
			"version",
		}),
		up: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "up",
			Help:      "1 if the op-dachallenger has finished starting up",
		}),
		challenges: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "challenges",
			Help:      "Number of challenges made by the challenge agent",
		}),
		failedChallenges: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "failed_challenges",
			Help:      "Number of challenges made by the challenge agent that have failed",
		}),
		resolutions: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "resolutions",
			Help:      "Number of resolutions made by the resolver agent",
		}),
		failedResolutions: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "failed_resolutions",
			Help:      "Number of resolutions made by the resolver agent that have failed",
		}),
		withdraws: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "withdraws",
			Help:      "Number of withdraws made by the withdraw agent",
		}),
		failedWithdraws: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "failed_withdraws",
			Help:      "Number of withdraws made by the withdraw agent that have failed",
		}),
		unlocks: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "unlocks",
			Help:      "Number of unlocks made by the unlock agent",
		}),
		failedUnlocks: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "failed_unlocks",
			Help:      "Number of unlocks made by the unlock agent that have failed",
		}),
		highestLastActedL1Block: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "highest_acted_l1_block",
			Help:      "Highest L1 block acted on by the da challenger service",
		}),
		recordedL1Ref: *factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "acted_l1_refs",
			Help:      "Tracks the L1 refs that have been acted on",
		}, []string{
			"block_hash",
		}),
	}
}

func (m *Metrics) Start(host string, port int) (*httputil.HTTPServer, error) {
	return opmetrics.StartServer(m.registry, host, port)
}

func (m *Metrics) StartBalanceMetrics(
	l log.Logger,
	client *ethclient.Client,
	account common.Address,
) io.Closer {
	return opmetrics.LaunchBalanceMetrics(l, m.registry, m.ns, client, account)
}

func (m *Metrics) Document() []opmetrics.DocumentedMetric {
	return m.factory.Document()
}

func (m *Metrics) RecordDAResolve() {
	m.resolutions.Add(1)
}

func (m *Metrics) RecordDAResolveFailed() {
	m.failedResolutions.Add(1)
}

func (m *Metrics) RecordDAChallenge() {
	m.challenges.Add(1)
}

func (m *Metrics) RecordDAChallengeFailed() {
	m.failedChallenges.Add(1)
}

func (m *Metrics) RecordWithdraw() {
	m.withdraws.Add(1)
}

func (m *Metrics) RecordWithdrawFailed() {
	m.failedWithdraws.Add(1)
}

func (m *Metrics) RecordBondUnlock() {
	m.unlocks.Add(1)
}

func (m *Metrics) RecordBondUnlockFailed() {
	m.failedUnlocks.Add(1)
}

func (m *Metrics) RecordLastActedL1Block(n uint64) {
	m.highestLastActedL1Block.Set(float64(n))
}

func (m *Metrics) RecordActedL1Ref(ref eth.L1BlockRef) {
	m.recordedL1Ref.WithLabelValues(ref.Hash.Hex()).Set(1)
}

func (m *Metrics) RecordUp() {
	m.up.Add(1)
}

func (m *Metrics) RecordInfo(version string) {
	m.info.WithLabelValues(version).Set(1)
}
