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
	"github.com/ethereum/go-ethereum/common"
)

type StakingMaster struct {
	Logger       logger.Logger
	ContractAddr common.Address

	Dispatcher dispatcher.DispatcherClaasic

	KeygenDoner core.KeygenDoner
	Storer      storage.MarshalSetter

	PChainIssueClient chain.Issuer

	keygenWatcher *KeygenRequestAddedEventWatcher
	keygenDealer  *KeygenRequestAddedEventHandler
}

func (s *StakingMaster) Start(_ context.Context) error {
	s.subscribe()
	return nil
}

func (s *StakingMaster) subscribe() {
	keygenWatcher := KeygenRequestAddedEventWatcher{
		Logger:       s.Logger,
		ContractAddr: s.ContractAddr,
		Publisher:    s.Dispatcher,
	}

	keygenDealer := KeygenRequestAddedEventHandler{
		Logger:      s.Logger,
		KeygenDoner: s.KeygenDoner,
		Storer:      s.Storer,
		Publisher:   s.Dispatcher,
	}

	s.keygenWatcher = &keygenWatcher
	s.keygenDealer = &keygenDealer

	s.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, s.keygenWatcher)
	s.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, s.keygenWatcher) // Emit event: *contract.MpcManagerKeygenRequestAdded

	s.Dispatcher.Subscribe(&events.GroupInfoStoredEvent{}, s.keygenDealer)
	s.Dispatcher.Subscribe(&contract.MpcManagerKeygenRequestAdded{}, s.keygenDealer) // Emit event: *events.GeneratedPubKeyInfoStoredEvent
}
