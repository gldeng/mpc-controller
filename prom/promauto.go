package prom

import (
	"github.com/alitto/pond"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	prefix = "mpc_controller_"
)

var (
	// Contract event metrics

	ContractEvtParticipantAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "contract_evt_participant_added_total",
		Help: "The total number of contract event ParticipantAdded",
	})

	ContractEvtKeygenRequestAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "contract_evt_keygen_request_added_total",
		Help: "The total number of contract event KeygenRequestAdded",
	})

	ContractEvtKeyGenerated = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "contract_evt_key_generated_total",
		Help: "The total number of contract event KeyGenerated",
	})

	ContractEvtStakeRequestAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "contract_evt_stake_request_added_total",
		Help: "The total number of contract event StakeRequestAdded",
	})

	ContractEvtRequestStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "contract_evt_request_started_total",
		Help: "The total number of contract event RequestStarted",
	})

	// Mpc keygen metrics

	MpcKeygenPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_keygen_posted_total",
		Help: "The total number of mpc keygen posted",
	})

	MpcKeygenDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_keygen_done_total",
		Help: "The total number of mpc keygen done",
	})

	// Mpc sign metrics

	MpcSignPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_posted_total",
		Help: "The total number of mpc sign posted",
	})

	MpcSignDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_done_total",
		Help: "The total number of mpc sign done",
	})

	MpcSignPostedForC2PExportTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_posted_total_for_c2p_export_tx",
		Help: "The total number of mpc sign posted for c2p ExportTx",
	})

	MpcSignDoneForC2PExportTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_done_total_for_c2p_export_tx",
		Help: "The total number of mpc sign done for c2p ExportTx",
	})

	MpcSignPostedForC2PImportTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_posted_total_for_c2p_import_tx",
		Help: "The total number of mpc sign posted for c2p ImportTx",
	})

	MpcSignDoneForC2PImportTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_done_total_for_c2p_import_tx",
		Help: "The total number of mpc sign done for c2p ImportTx",
	})

	MpcSignPostedForAddDelegatorTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_posted_total_for_add_delegator_tx",
		Help: "The total number of mpc sign posted for AddDelegatorx",
	})

	MpcSignDoneForAddDelegatorTx = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_sign_done_total_for_add_delegator_tx",
		Help: "The total number of mpc sign done for AddDelegatorTx",
	})

	// Mpc result metrics

	MpcResultPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_result_posted_total",
		Help: "The total number of mpc result posted",
	})
	MpcResulDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_result_done_total",
		Help: "The total number of mpc result done",
	})

	// Mpc join metrics

	MpcJoinStake = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_join_stake_total",
		Help: "The total number of mpc join stake",
	})

	MpcJoinStakeQuorumReached = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_join_stake_quorum_reached_total",
		Help: "The total number of mpc join stake quorum reached",
	})

	// Issue tx metrics

	C2PExportTxIssued = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "c2p_export_tx_issued_total",
		Help: "The total number of c2p ExportTx issued",
	})

	C2PExportTxCommitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "c2p_export_tx_committed_total",
		Help: "The total number of c2p ExportTx committed",
	})

	C2PImportTxIssued = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "cp2_import_tx_issued_total",
		Help: "The total number of cp2 ImportTx issued",
	})

	C2PImportTxCommitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "cp2_import_tx_committed_total",
		Help: "The total number of cp2 ImportTx committed",
	})

	AddDelegatorTxIssued = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "add_delegator_tx_issued_total",
		Help: "The total number of AddDelegatorTx issued",
	})

	AddDelegatorTxCommitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "add_delegator_tx_committed_total",
		Help: "The total number of AddDelegatorTx committed",
	})
)

// Reference: https://github.com/alitto/pond

func ConfigWorkPoolAndTaskMetrics(poolType string, pool *pond.WorkerPool) {
	// Worker pool metrics
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_workers_running",
			Help: "Number of running worker goroutines",
		},
		func() float64 {
			return float64(pool.RunningWorkers())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_workers_idle",
			Help: "Number of idle worker goroutines",
		},
		func() float64 {
			return float64(pool.IdleWorkers())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_workers_minimum",
			Help: "Minimum number of worker goroutines",
		},
		func() float64 {
			return float64(pool.MinWorkers())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_workers_maxmimum",
			Help: "Maxmimum number of worker goroutines",
		},
		func() float64 {
			return float64(pool.MaxWorkers())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_queue_capacity",
			Help: "Maximum number of tasks that can be waiting in the queue at any given time (queue capacity)",
		},
		func() float64 {
			return float64(pool.MaxCapacity())
		}))

	// Task metrics
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: prefix + poolType + "pool_tasks_submitted_total",
			Help: "Number of tasks submitted",
		},
		func() float64 {
			return float64(pool.SubmittedTasks())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: prefix + poolType + "pool_tasks_waiting_total",
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 {
			return float64(pool.WaitingTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: prefix + poolType + "pool_tasks_successful_total",
			Help: "Number of tasks that completed successfully",
		},
		func() float64 {
			return float64(pool.SuccessfulTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: prefix + poolType + "pool_tasks_failed_total",
			Help: "Number of tasks that completed with panic",
		},
		func() float64 {
			return float64(pool.FailedTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: prefix + poolType + "pool_tasks_completed_total",
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 {
			return float64(pool.CompletedTasks())
		}))
}
