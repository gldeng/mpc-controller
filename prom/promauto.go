package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	prefix = "mpc_controller_"
)

var (
	// Contract events

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

	// Mpc keygen

	MpcKeygenPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_keygen_posted_total",
		Help: "The total number of mpc keygen posted",
	})

	MpcKeygenDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_keygen_done_total",
		Help: "The total number of mpc keygen done",
	})

	// Mpc sign

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

	// Mpc result

	MpcResultPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_result_posted_total",
		Help: "The total number of mpc result posted",
	})
	MpcResulDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_result_done_total",
		Help: "The total number of mpc result done",
	})

	// Mpc join

	MpcJoinStake = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_join_stake_total",
		Help: "The total number of mpc join stake",
	})

	MpcJoinStakeQuorumReached = promauto.NewCounter(prometheus.CounterOpts{
		Name: prefix + "mpc_join_stake_quorum_reached_total",
		Help: "The total number of mpc join stake quorum reached",
	})

	// Issue tx

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
