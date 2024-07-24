package metrics

import (
	"io"

	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum-optimism/optimism/op-service/httputil"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"

	contractMetrics "github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	txmetrics "github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
)

const Namespace = "op_challenger"

type Metricer interface {
	RecordInfo(version string)
	RecordUp()

	StartBalanceMetrics(l log.Logger, client *ethclient.Client, account common.Address) io.Closer

	// Record Tx metrics
	txmetrics.TxMetricer

	// Record cache metrics
	caching.Metrics

	// Record contract metrics
	contractMetrics.ContractMetricer

	RecordActedL1Block(n uint64)

	RecordDAChallenge()
	RecordDAChallengeResolve()
	RecordDAChallengeSucceed()
	RecordDAChallengeFailed()
	RecordDAChallengeResolutionTime(t float64)
	RecordDAChallengeTime(t float64)

	RecordBondUnlockFailed()

	RecordWithdrawFailed()

	RecordGamesStatus(inProgress, defenderWon, challengerWon int)

	IncActiveExecutors()
	DecActiveExecutors()
	IncIdleExecutors()
	DecIdleExecutors()

	RecordL1ReorgDepth(d uint64)
	RecordL1Ref(name string, ref eth.L1BlockRef)
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

	info prometheus.GaugeVec
	up   prometheus.Gauge

	executors prometheus.GaugeVec

	bondClaimFailures prometheus.Counter
	bondsClaimed      prometheus.Counter

	highestActedL1Block prometheus.Gauge

	challenges  prometheus.Counter
	resolutions prometheus.Counter
	succeeded   prometheus.Counter
	failed      prometheus.Counter

	resolutionTime prometheus.Histogram
	challengeTime  prometheus.Histogram

	trackedGames  prometheus.GaugeVec
	inflightGames prometheus.Gauge
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
			Help:      "1 if the op-challenger has finished starting up",
		}),
		executors: *factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "executors",
			Help:      "Number of active and idle executors",
		}, []string{
			"status",
		}),
		challenges: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "challenges",
			Help:      "Number of challenges made by the challenge agent",
		}),
		resolutions: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "resolutions",
			Help:      "Number of resolutions made by the resolver agent",
		}),
		succeeded: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "succeeded",
			Help:      "Number of successful challenges",
		}),
		failed: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "failed",
			Help:      "Number of failed challenges",
		}),
		resolutionTime: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: Namespace,
			Name:      "resolution_time",
			Help:      "Time (in seconds) spent trying to resolve challenges",
			Buckets:   []float64{.05, .1, .25, .5, 1, 2.5, 5, 7.5, 10},
		}),
		challengeTime: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: Namespace,
			Name:      "challenge_time",
			Help:      "Time (in seconds) spent submitting challenges",
			Buckets:   []float64{.05, .1, .25, .5, 1, 2.5, 5, 7.5, 10},
		}),
		bondClaimFailures: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "claim_failures",
			Help:      "Number of bond claims that failed",
		}),
		bondsClaimed: factory.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "bonds",
			Help:      "Number of bonds claimed by the challenge agent",
		}),
		trackedGames: *factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "tracked_games",
			Help:      "Number of games being tracked by the challenger",
		}, []string{
			"status",
		}),
		highestActedL1Block: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "highest_acted_l1_block",
			Help:      "Highest L1 block acted on by the challenger",
		}),
		inflightGames: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "inflight_games",
			Help:      "Number of games being tracked by the challenger",
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

// RecordInfo sets a pseudo-metric that contains versioning and
// config info for the op-proposer.
func (m *Metrics) RecordInfo(version string) {
	m.info.WithLabelValues(version).Set(1)
}

// RecordUp sets the up metric to 1.
func (m *Metrics) RecordUp() {
	prometheus.MustRegister()
	m.up.Set(1)
}

func (m *Metrics) RecordL1ReorgDepth(d uint64) {
	//TODO implement me
	panic("implement me")
}

func (m *Metrics) RecordL1Ref(name string, ref eth.L1BlockRef) {
	//TODO implement me
	panic("implement me")
}

func (m *Metrics) Document() []opmetrics.DocumentedMetric {
	return m.factory.Document()
}

func (m *Metrics) RecordDAChallenge() {
	panic("implement me")
}

func (m *Metrics) RecordDAChallengeResolve() {
	panic("implement me")
}

func (m *Metrics) RecordDAChallengeSucceed() {
	panic("implement me")
}

func (m *Metrics) RecordDAChallengeFailed() {
	panic("implement me")
}

func (m *Metrics) RecordDAChallengeResolutionTime(t float64) {
	panic("implement me")
}

func (m *Metrics) RecordDAChallengeTime(t float64) {
	panic("implement me")
}

func (m *Metrics) RecordWithdrawFailed() {
	//TODO implement me
	panic("implement me")
}

func (m *Metrics) RecordBondUnlockFailed() {
	//TODO implement me
	panic("implement me")
}

func (m *Metrics) RecordBondClaimFailed() {
	m.bondClaimFailures.Add(1)
}

func (m *Metrics) RecordBondClaimed(amount uint64) {
	m.bondsClaimed.Add(float64(amount))
}

func (m *Metrics) IncActiveExecutors() {
	m.executors.WithLabelValues("active").Inc()
}

func (m *Metrics) DecActiveExecutors() {
	m.executors.WithLabelValues("active").Dec()
}

func (m *Metrics) IncIdleExecutors() {
	m.executors.WithLabelValues("idle").Inc()
}

func (m *Metrics) DecIdleExecutors() {
	m.executors.WithLabelValues("idle").Dec()
}

func (m *Metrics) RecordGamesStatus(inProgress, defenderWon, challengerWon int) {
	m.trackedGames.WithLabelValues("in_progress").Set(float64(inProgress))
	m.trackedGames.WithLabelValues("defender_won").Set(float64(defenderWon))
	m.trackedGames.WithLabelValues("challenger_won").Set(float64(challengerWon))
}

func (m *Metrics) RecordActedL1Block(n uint64) {
	m.highestActedL1Block.Set(float64(n))
}

func (m *Metrics) RecordGameUpdateScheduled() {
	m.inflightGames.Add(1)
}

func (m *Metrics) RecordGameUpdateCompleted() {
	m.inflightGames.Sub(1)
}
