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
	"github.com/avalido/mpc-controller/tasks/rewarding/export"
	"github.com/avalido/mpc-controller/tasks/rewarding/report"
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
	utxoFetcher           *report.UTXOFetcher
	rewardedStakeReporter *report.RewardedStakeReporter

	// export
	exportRewardReqAddedEvtWatcher  *export.ExportRewardRequestAddedEventWatcher
	exportRewardJoiner              *export.ExportRewardRequestJoiner
	exportRewarReqStartedEvtWatcher *export.ExportRewardRequestStartedEventWatcher
	rewardExporter                  *export.StakingRewardExporter
}

func (m *RewardingMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *RewardingMaster) subscribe() {
	utxoFetcher := report.UTXOFetcher{
		Logger:       m.Logger,
		PChainClient: m.PChainClient,
		Publisher:    m.Dispatcher,
	}

	rewardedStakeReporter := report.RewardedStakeReporter{
		Cache:           m.Cache,
		ContractAddr:    m.ContractAddr,
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Publisher:       m.Dispatcher,
		Receipter:       m.Receipter,
		Signer:          m.Signer,
		Transactor:      m.Transactor,
	}

	exportRewardReqAddedEvtWatcher := export.ExportRewardRequestAddedEventWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	exportRewardJoiner := export.ExportRewardRequestJoiner{
		Cache:           m.Cache,
		ContractAddr:    m.ContractAddr,
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Publisher:       m.Dispatcher,
		Receipter:       m.Receipter,
		Signer:          m.Signer,
		Transactor:      m.Transactor,
	}

	exportRewarReqStartedEvtWatcher := export.ExportRewardRequestStartedEventWatcher{
		ContractAddr: m.ContractAddr,
		Logger:       m.Logger,
		Publisher:    m.Dispatcher,
	}

	rewardExporter := export.StakingRewardExporter{
		CChainIssueClient: m.CChainIssueClient,
		Cache:             m.Cache,
		Logger:            m.Logger,
		MyPubKeyHashHex:   m.MyPubKeyHashHex,
		NetworkContext:    m.NetworkContext,
		PChainIssueClient: m.PChainClient,
		Publisher:         m.Dispatcher,
		SignDoner:         m.SignDoner,
	}

	m.utxoFetcher = &utxoFetcher
	m.rewardedStakeReporter = &rewardedStakeReporter

	m.exportRewardReqAddedEvtWatcher = &exportRewardReqAddedEvtWatcher
	m.exportRewardJoiner = &exportRewardJoiner
	m.exportRewarReqStartedEvtWatcher = &exportRewarReqStartedEvtWatcher
	m.rewardExporter = &rewardExporter

	m.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, m.utxoFetcher)        // Emit event: *events.StakingPeriodEndedEvent
	m.Dispatcher.Subscribe(&events.UTXOsFetchedEvent{}, m.rewardedStakeReporter) // Emit event: *events.RewardedStakeReportedEvent

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportRewardReqAddedEvtWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.exportRewardReqAddedEvtWatcher) // Emit event: *contract.ExportRewardRequestAddedEvent
	m.Dispatcher.Subscribe(&events.ExportRewardRequestAddedEvent{}, m.exportRewardJoiner)              // Emit event: *events.JoinedExportRewardRequestEvent
	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.exportRewarReqStartedEvtWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.exportRewarReqStartedEvtWatcher) // Emit event: *contract.ExportRewardRequestStartedEvent
	m.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, m.rewardExporter)
	m.Dispatcher.Subscribe(&events.UTXOsFetchedEvent{}, m.rewardExporter)
	m.Dispatcher.Subscribe(&events.ExportRewardRequestStartedEvent{}, m.rewardExporter) // Emit event: *events.RewardExportedEvent
}
