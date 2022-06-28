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
	Logger       logger.Logger
	ContractAddr common.Address

	MyPubKeyHashHex string

	Dispatcher dispatcher.DispatcherClaasic

	chain.NetworkContext

	Cache Cache

	SignDoner core.SignDoner

	Noncer chain.Noncer

	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer

	stakingWatcher *StakeRequestStartedEventWatcher
	stakingDealer  *StakeRequestStartedEventHandler
}

func (s *StakingMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *StakingMaster) subscribe() {
	taskStartedWatcher := StakeRequestStartedEventWatcher{
		Logger:       s.Logger,
		ContractAddr: s.ContractAddr,
		Publisher:    s.Dispatcher,
	}

	taskStartedDealer := StakeRequestStartedEventHandler{
		Logger:            s.Logger,
		NetworkContext:    s.NetworkContext,
		MyPubKeyHashHex:   s.MyPubKeyHashHex,
		Cache:             s.Cache,
		SignDoner:         s.SignDoner,
		Publisher:         s.Dispatcher,
		CChainIssueClient: s.CChainIssueClient,
		PChainIssueClient: s.PChainIssueClient,
		Noncer:            s.Noncer,
	}

	s.stakingWatcher = &taskStartedWatcher
	s.stakingDealer = &taskStartedDealer

	s.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, s.stakingWatcher)
	s.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, s.stakingWatcher) // Emit event: *contract.MpcManagerStakeRequestStarted

	s.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestStarted{}, s.stakingDealer) // Emit event: *events.StakingTaskDoneEvent
}
