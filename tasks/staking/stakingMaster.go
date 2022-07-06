package staking

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
)

type StakingMaster struct {
	CChainIssueClient chain.CChainIssuer
	Cache             Cache
	ContractAddr      common.Address
	Dispatcher        dispatcher.DispatcherClaasic
	Logger            logger.Logger
	MyPubKeyHashHex   string
	Noncer            chain.Noncer
	PChainIssueClient chain.PChainIssuer
	SignDoner         core.SignDoner
	chain.NetworkContext

	stakingWatcher *StakeRequestStartedEventWatcher
	stakingDealer  *StakeRequestStartedEventHandler
}

func (m *StakingMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *StakingMaster) subscribe() {
	taskStartedWatcher := StakeRequestStartedEventWatcher{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Publisher:    m.Dispatcher,
	}

	taskStartedDealer := StakeRequestStartedEventHandler{
		Logger:            m.Logger,
		NetworkContext:    m.NetworkContext,
		MyPubKeyHashHex:   m.MyPubKeyHashHex,
		Cache:             m.Cache,
		SignDoner:         m.SignDoner,
		Publisher:         m.Dispatcher,
		CChainIssueClient: m.CChainIssueClient,
		PChainIssueClient: m.PChainIssueClient,
		Noncer:            m.Noncer,
	}

	m.stakingWatcher = &taskStartedWatcher
	m.stakingDealer = &taskStartedDealer

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.stakingWatcher)
	m.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, m.stakingWatcher) // Emit event: *contract.MpcManagerStakeRequestStarted

	m.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestStarted{}, m.stakingDealer) // Emit event: *events.StakingTaskDoneEvent
}
