package staking

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
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
	Verifyier crypto.VerifyHasher

	Noncer chain.Noncer

	CChainIssueClient chain.Issuer
	PChainIssueClient chain.Issuer

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

	issuer := Issuer{
		Logger:            s.Logger,
		CChainIssueClient: s.CChainIssueClient,
		PChainIssueClient: s.PChainIssueClient,
	}

	taskStartedDealer := StakeRequestStartedEventHandler{
		Logger:          s.Logger,
		NetworkContext:  s.NetworkContext,
		MyPubKeyHashHex: s.MyPubKeyHashHex,
		Cache:           s.Cache,
		SignDoner:       s.SignDoner,
		Verifyier:       s.Verifyier,
		Noncer:          s.Noncer,
		Issuer:          &issuer,
	}

	s.stakingWatcher = &taskStartedWatcher
	s.stakingDealer = &taskStartedDealer

	s.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, s.stakingWatcher)
	s.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, s.stakingWatcher) // Emit event: *contract.MpcManagerStakeRequestStarted

	s.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestStarted{}, s.stakingDealer)
}
