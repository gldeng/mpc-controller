package rewarding

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tasks/rewarding/joiner"
	"github.com/avalido/mpc-controller/tasks/rewarding/porter"
	"github.com/avalido/mpc-controller/tasks/rewarding/tracker"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type RewardingMaster struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	ContractAddr      common.Address
	Dispatcher        dispatcher.DispatcherClaasic
	Logger            logger.Logger
	MyPubKeyHashHex   string
	PChainClient      platformvm.Client
	Receipter         chain.Receipter
	RewardUTXOGetter  chain.RewardUTXOGetter
	SignDoner         core.SignDoner
	Signer            *bind.TransactOpts
	Transactor        bind.ContractTransactor
	chain.NetworkContext

	// report
	utxoTracker *tracker.UTXOTracker

	// export
	exportRewardReqAddedEvtWatcher  *joiner.ExportRewardRequestAddedEventWatcher
	exportRewardJoiner              *joiner.ExportUTXORequestJoiner
	exportRewarReqStartedEvtWatcher *porter.ExportRewardRequestStartedEventWatcher
	rewardExporter                  *porter.StakingRewardExporter
}

func (m *RewardingMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *RewardingMaster) subscribe() {
	utxoFetcher := tracker.UTXOTracker{
		Logger:       m.Logger,
		PChainClient: m.PChainClient,
		Publisher:    m.Dispatcher,
	}

	exportRewardReqAddedEvtWatcher := joiner.ExportRewardRequestAddedEventWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	exportRewardJoiner := joiner.ExportUTXORequestJoiner{
		Cache:           m.Cache,
		ContractAddr:    m.ContractAddr,
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Publisher:       m.Dispatcher,
		Receipter:       m.Receipter,
		Signer:          m.Signer,
		Transactor:      m.Transactor,
	}

	exportRewarReqStartedEvtWatcher := porter.ExportRewardRequestStartedEventWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	rewardExporter := porter.StakingRewardExporter{
		CChainIssueClient: m.CChainIssueClient,
		Cache:             m.Cache,
		Logger:            m.Logger,
		MyPubKeyHashHex:   m.MyPubKeyHashHex,
		NetworkContext:    m.NetworkContext,
		PChainIssueClient: m.PChainClient,
		Publisher:         m.Dispatcher,
		SignDoner:         m.SignDoner,
	}

	m.utxoTracker = &utxoFetcher

	m.exportRewardReqAddedEvtWatcher = &exportRewardReqAddedEvtWatcher
	m.exportRewardJoiner = &exportRewardJoiner
	m.exportRewarReqStartedEvtWatcher = &exportRewarReqStartedEvtWatcher
	m.rewardExporter = &rewardExporter

	m.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, m.utxoTracker) // Emit event: *events.StakingPeriodEndedEvent

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportRewardReqAddedEvtWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.exportRewardReqAddedEvtWatcher) // Emit event: *contract.ExportUTXORequestAddedEvent
	m.Dispatcher.Subscribe(&events.ExportUTXORequestAddedEvent{}, m.exportRewardJoiner)                // Emit event: *events.JoinedExportUTXORequestEvent
	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportRewarReqStartedEvtWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.exportRewarReqStartedEvtWatcher) // Emit event: *contract.ExportUTXORequestStartedEvent
	m.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, m.rewardExporter)
	m.Dispatcher.Subscribe(&events.UTXOsFetchedEvent{}, m.rewardExporter)
	m.Dispatcher.Subscribe(&events.ExportUTXORequestStartedEvent{}, m.rewardExporter) // Emit event: *events.UTXOExportedEvent
}
