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

type StakingService struct {
	Logger       logger.Logger
	ContractAddr common.Address

	MyPubKeyHashHex string

	Dispatcher dispatcher.DispatcherClaasic

	chain.NetworkContext

	Storer Storer

	SignDoner core.SignDoner
	Verifyier crypto.VerifyHasher

	Noncer chain.Noncer

	CChainIssueClient chain.Issuer
	PChainIssueClient chain.Issuer
}

func (s *StakingService) Start(_ context.Context) error {
	s.subscribe()
	return nil
}

func (s *StakingService) subscribe() {
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
		Storer:          s.Storer,
		SignDoner:       s.SignDoner,
		Verifyier:       s.Verifyier,
		Noncer:          s.Noncer,
		Issuer:          &issuer,
	}

	s.Dispatcher.Subscribe(&events.ContractFiltererReconnectedEvent{}, &taskStartedWatcher)
	s.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, &taskStartedWatcher) // Emit event: *contract.MpcManagerStakeRequestStarted

	s.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestStarted{}, &taskStartedDealer)
}
