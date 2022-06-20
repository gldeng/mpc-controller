package keygen

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type KeygenMaster struct {
	Logger       logger.Logger
	ContractAddr common.Address

	MyPubKeyHashHex string

	Dispatcher dispatcher.DispatcherClaasic

	KeygenDoner core.KeygenDoner
	Storer      storage.MarshalSetter

	Transactor bind.ContractTransactor
	Signer     *bind.TransactOpts
	Receipter  chain.Receipter

	PChainIssueClient chain.Issuer

	keygenWatcher *KeygenRequestAddedEventWatcher
	keygenDealer  *KeygenRequestAddedEventHandler
}

func (s *KeygenMaster) Start(ctx context.Context) error {
	s.subscribe()
	<-ctx.Done()
	return nil
}

func (s *KeygenMaster) subscribe() {
	keygenWatcher := KeygenRequestAddedEventWatcher{
		Logger:       s.Logger,
		ContractAddr: s.ContractAddr,
		Publisher:    s.Dispatcher,
	}

	keygenDealer := KeygenRequestAddedEventHandler{
		Logger:          s.Logger,
		KeygenDoner:     s.KeygenDoner,
		Storer:          s.Storer,
		Publisher:       s.Dispatcher,
		MyPubKeyHashHex: s.MyPubKeyHashHex,
		Transactor:      s.Transactor,
		Signer:          s.Signer,
		Receipter:       s.Receipter,
	}

	s.keygenWatcher = &keygenWatcher
	s.keygenDealer = &keygenDealer

	s.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, s.keygenWatcher)
	s.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, s.keygenWatcher) // Emit event: *contract.MpcManagerKeygenRequestAdded

	s.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, s.keygenDealer)
	s.Dispatcher.Subscribe(&contract.MpcManagerKeygenRequestAdded{}, s.keygenDealer) // Emit event: *events.GeneratedPubKeyInfoStoredEvent
}
