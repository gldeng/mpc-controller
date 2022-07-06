package rewarding

import (
	"context"
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
	Logger           logger.Logger
	ContractAddr     common.Address
	RewardUTXOGetter chain.RewardUTXOGetter
	Dispatcher       dispatcher.DispatcherClaasic
	MyPubKeyHashHex  string
	Cache            cache.ICache
	SignDoner        core.SignDoner
	Signer           *bind.TransactOpts
	Transactor       bind.ContractTransactor
	Receipter        chain.Receipter
	chain.NetworkContext
	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer

	// report
	periodEndedChecker    *report.StakingPeriodEndedChecker
	rewardUTXOFetcher     *report.StakingRewardUTXOFetcher
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
	periodEndedChecker := report.StakingPeriodEndedChecker{
		Publisher: m.Dispatcher,
	}

	rewardUTXOFetcher := report.StakingRewardUTXOFetcher{
		Logger:           m.Logger,
		Publisher:        m.Dispatcher,
		RewardUTXOGetter: m.RewardUTXOGetter,
	}

	rewardedStakeReporter := report.RewardedStakeReporter{
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Publisher:       m.Dispatcher,
		Cache:           m.Cache,
		Signer:          m.Signer,
		ContractAddr:    m.ContractAddr,
		Transactor:      m.Transactor,
		Receipter:       m.Receipter,
	}

	exportRewardReqAddedEvtWatcher := export.ExportRewardRequestAddedEventWatcher{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Publisher:    m.Dispatcher,
	}

	exportRewardJoiner := export.ExportRewardRequestJoiner{
		Logger:          m.Logger,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		Signer:          m.Signer,
		Cache:           m.Cache,
		Transactor:      m.Transactor,
		ContractAddr:    m.ContractAddr,
		Publisher:       m.Dispatcher,
		Receipter:       m.Receipter,
	}

	exportRewarReqStartedEvtWatcher := export.ExportRewardRequestStartedEventWatcher{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Publisher:    m.Dispatcher,
	}

	rewardExporter := export.StakingRewardExporter{
		Logger:            m.Logger,
		NetworkContext:    m.NetworkContext,
		MyPubKeyHashHex:   m.MyPubKeyHashHex,
		Publisher:         m.Dispatcher,
		SignDoner:         m.SignDoner,
		CChainIssueClient: m.CChainIssueClient,
		PChainIssueClient: m.PChainIssueClient,
		Cache:             m.Cache,
	}

	m.periodEndedChecker = &periodEndedChecker
	m.rewardUTXOFetcher = &rewardUTXOFetcher
	m.rewardedStakeReporter = &rewardedStakeReporter

	m.exportRewardReqAddedEvtWatcher = &exportRewardReqAddedEvtWatcher
	m.exportRewardJoiner = &exportRewardJoiner
	m.exportRewarReqStartedEvtWatcher = &exportRewarReqStartedEvtWatcher
	m.rewardExporter = &rewardExporter

	m.Dispatcher.Subscribe(&events.StakingTaskDoneEvent{}, m.periodEndedChecker) // Emit event: *events.StakingPeriodEndedEvent

	m.Dispatcher.Subscribe(&events.StakingPeriodEndedEvent{}, m.rewardUTXOFetcher) // Emit event: *events.RewardUTXOsFetchedEvent
}
