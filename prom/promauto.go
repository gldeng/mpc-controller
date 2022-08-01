package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	prefix = "mpc_"
)

var (
	// stake

	StakeRequestAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "stake_request_added_total",
		Help: "The total number of StakeRequestAddedEvent",
	})
	StakeTaskDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "stake_task_done_total",
		Help: "The total number of StakeTaskDoneEvent",
	})

	// reward

	PrincipalUTXOExported = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "principal_utxo_exported_total",
		Help: "The total number of PrincipalUTXOExportedEvent",
	})
	RewardUTXOExported = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "reward_utxo_exported_total",
		Help: "The total number of RewardUTXOExportedEvent",
	})

	// sign task added

	StakeSignTaskAdded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "stake_sign_task_added_total",
		Help: "The total number of stake sign task added",
	})

	PrincipalUTXOSignTaskAdded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "principal_utxo_sign_task_added_total",
		Help: "The total number of principal UTXO sign task added",
	})

	RewardUTXOSignTaskAdded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "reward_utxo_sign_task_added_total",
		Help: "The total number of reward UTXO sign task added",
	})

	// sign task done

	StakeSignTaskDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "stake_sign_task_done_total",
		Help: "The total number of of stake sign task done",
	})

	PrincipalUTXOSignTaskDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "principal_utxo_sign_task_done_total",
		Help: "The total number of of principal UTXO sign task done",
	})

	RewardUTXOSignTaskDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "reward_utxo_sign_task_done_total",
		Help: "The total number of reward UTXO sign task done",
	})
)
