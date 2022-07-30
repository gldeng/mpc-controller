package participant

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ParticipantMaster struct {
	ContractAddr    common.Address
	ContractCaller  bind.ContractCaller
	Dispatcher      dispatcher.Dispatcher
	Logger          logger.Logger
	MyPubKeyBytes   []byte
	MyPubKeyHashHex string
	MyPubKeyHex     string
	Storer          storage.MarshalSetter

	partiEvtWatcher *ParticipantAddedEventWatcher
	partiInfoStorer *ParticipantInfoStorer
	groupInfoStorer *GroupInfoStorer
}

func (m *ParticipantMaster) Start(ctx context.Context) error {
	m.subscribe()
	<-ctx.Done()
	return nil
}

func (m *ParticipantMaster) subscribe() {
	partiWatcher := ParticipantAddedEventWatcher{
		Logger:        m.Logger,
		MyPubKeyBytes: m.MyPubKeyBytes,
		ContractAddr:  m.ContractAddr,
		Publisher:     m.Dispatcher,
	}

	partiInfoStorer := ParticipantInfoStorer{
		Logger:          m.Logger,
		Publisher:       m.Dispatcher,
		Storer:          m.Storer,
		MyPubKeyHex:     m.MyPubKeyHex,
		MyPubKeyHashHex: m.MyPubKeyHashHex,
	}

	groupInfoStorer := GroupInfoStorer{
		Logger:       m.Logger,
		ContractAddr: m.ContractAddr,
		Caller:       m.ContractCaller,
		Publisher:    m.Dispatcher,
		Storer:       m.Storer,
	}

	m.partiEvtWatcher = &partiWatcher
	m.partiInfoStorer = &partiInfoStorer
	m.groupInfoStorer = &groupInfoStorer

	m.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, m.partiEvtWatcher) // Emit event: *contract.MpcManagerParticipantAdded

	m.Dispatcher.Subscribe(&contract.MpcManagerParticipantAdded{}, m.partiInfoStorer) // Emit event: *events.ParticipantInfoStoredEvent
	m.Dispatcher.Subscribe(&contract.MpcManagerParticipantAdded{}, m.groupInfoStorer) // Emit event: *events.GroupInfoStoredEvent
}
