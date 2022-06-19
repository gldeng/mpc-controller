package participant

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type ParticipantMaster struct {
	Logger logger.Logger

	MyPubKeyHex     string
	MyPubKeyHashHex string
	MyPubKeyBytes   []byte

	ContractAddr   common.Address
	ContractCaller bind.ContractCaller

	Dispatcher dispatcher.DispatcherClaasic
	Storer     storage.MarshalSetter

	partiEvtWatcher *ParticipantAddedEventWatcher
	partiInfoStorer *ParticipantInfoStorer
	groupInfoStorer *GroupInfoStorer
}

func (p *ParticipantMaster) Start(ctx context.Context) error {
	fmt.Printf("ParticipantMaster enered at: %v", time.Now())

	p.subscribe()

	<-ctx.Done()
	fmt.Printf("ParticipantMaster exited at: %v", time.Now())

	return nil
}

func (p *ParticipantMaster) subscribe() {
	partiWatcher := ParticipantAddedEventWatcher{
		Logger:        p.Logger,
		MyPubKeyBytes: p.MyPubKeyBytes,
		ContractAddr:  p.ContractAddr,
		Publisher:     p.Dispatcher,
	}

	partiInfoStorer := ParticipantInfoStorer{
		Logger:          p.Logger,
		Publisher:       p.Dispatcher,
		Storer:          p.Storer,
		MyPubKeyHex:     p.MyPubKeyHex,
		MyPubKeyHashHex: p.MyPubKeyHashHex,
	}

	groupInfoStorer := GroupInfoStorer{
		Logger:       p.Logger,
		ContractAddr: p.ContractAddr,
		Caller:       p.ContractCaller,
		Publisher:    p.Dispatcher,
		Storer:       p.Storer,
	}

	p.partiEvtWatcher = &partiWatcher
	p.partiInfoStorer = &partiInfoStorer
	p.groupInfoStorer = &groupInfoStorer

	p.Dispatcher.Subscribe(&events.ContractFiltererCreatedEvent{}, p.partiEvtWatcher) // Emit event: *contract.MpcManagerParticipantAdded

	p.Dispatcher.Subscribe(&contract.MpcManagerParticipantAdded{}, p.partiInfoStorer) // Emit event: *events.ParticipantInfoStoredEvent
	p.Dispatcher.Subscribe(&contract.MpcManagerParticipantAdded{}, p.groupInfoStorer) // Emit event: *events.GroupInfoStoredEvent
}
