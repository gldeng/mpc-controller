package joining

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type JoiningService struct {
	Logger       logger.Logger
	ContractAddr common.Address

	MyPubKeyHashHex string

	Dispatcher dispatcher.DispatcherClaasic

	Storer Storer

	Signer *bind.TransactOpts

	Receipter chain.Receipter

	Transactor bind.ContractTransactor
}

func (s *JoiningService) Start(_ context.Context) error {
	s.subscribe()
	return nil
}

func (s *JoiningService) subscribe() {
	stakeAddedWatcher := StakeRequestAddedEventWatcher{
		Logger:       s.Logger,
		ContractAddr: s.ContractAddr,
		Publisher:    s.Dispatcher,
	}

	stakeAddedHandler := StakeRequestAddedEventHandler{
		Logger:          s.Logger,
		MyPubKeyHashHex: s.MyPubKeyHashHex,
		Signer:          s.Signer,
		Storer:          s.Storer,
		Receipter:       s.Receipter,
		ContractAddr:    s.ContractAddr,
		Transactor:      s.Transactor,
		Publisher:       s.Dispatcher,
	}

	s.Dispatcher.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, &stakeAddedWatcher) // Emit event:  *contract.MpcManagerStakeRequestAdded
	s.Dispatcher.Subscribe(&contract.MpcManagerStakeRequestAdded{}, &stakeAddedHandler)  // Emit event: *events.JoinedRequestEvent
}
