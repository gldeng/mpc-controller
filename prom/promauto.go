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
	StakeRequestStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "stake_request_started_total",
		Help: "The total number of StakeRequestStartedEvent",
	})
	StakeTaskDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "stake_task_done_total",
		Help: "The total number of StakeTaskDoneEvent",
	})

	// reward

	UTXOExportRequestJoined = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "utxo_export_request_joined_total",
		Help: "The total number of UTXO export request joined",
	})
	PrincipalUTXOExportRequestJoined = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "principal_utxo_export_request_joined_total",
		Help: "The total number of principal UTXO export request joined",
	})
	RewardUTXOExportRequestJoined = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "reward_utxo_export_request_joined_total",
		Help: "The total number of reward UTXO export request joined",
	})
	UTXOExportRequestStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "utxo_export_request_started_total",
		Help: "The total number of UTXO export request started",
	})
	UTXOExported = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "utxo_exported_total",
		Help: "The total number of UTXO exported",
	})
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

var (
	DispatcherPublishedEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "dispatcher_published_events_total",
		Help: "The total number of dispatcher published events",
	})

	WorkshopWorkspaces = promauto.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "workshop_workspaces_total",
		Help: "The total number of workshop workspaces",
	})
)
