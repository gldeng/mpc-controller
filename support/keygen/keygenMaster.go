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

	keygenWatcher *KeygenRequestAddedEventWatcher
	keygenDealer  *KeygenRequestAddedEventHandler
}

func (m *KeygenMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *KeygenMaster) subscribe() {
	keygenWatcher := KeygenRequestAddedEventWatcher{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Publisher:    m.Dispatcher,
	}

	keygenDealer := KeygenRequestAddedEventHandler{
		Logger:          m.Logger,
		KeygenDoner:     m.KeygenDoner,
		Storer:          m.Storer,
		Publisher:       m.Dispatcher,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
		ContractAddr:    m.ContractAddr,
		Transactor:      m.Transactor,
		Signer:          m.Signer,
		Receipter:       m.Receipter,
	}

	m.keygenWatcher = &keygenWatcher
	m.keygenDealer = &keygenDealer

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.keygenWatcher)
	m.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, m.keygenWatcher) // Emit event: *contract.MpcManagerKeygenRequestAdded

	m.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, m.keygenDealer)
	m.Dispatcher.Subscribe(&events.ParticipantInfoStoredEvent{}, m.keygenDealer)
	m.Dispatcher.Subscribe(&contract.MpcManagerKeygenRequestAdded{}, m.keygenDealer) // Emit event: *events.GeneratedPubKeyInfoStoredEvent, *events.ReportedGenPubKeyEvent
}
