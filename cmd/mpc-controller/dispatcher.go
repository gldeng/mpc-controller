package main

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	contractWatchers "github.com/avalido/mpc-controller/contract/watchers"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/queue"
	storageHandlers "github.com/avalido/mpc-controller/storage/handlers"
)

func Dispatcher(ctx context.Context, logger logger.Logger) *dispatcher.Dispatcher {
	// Create New dispatcher
	d := dispatcher.NewDispatcher(ctx, logger, queue.NewArrayQueue(1024), 1024)

	// Subscribe events upon configurations
	d.Subscribe(&events.MpcControllerPubKeyConfiguredEvent{}, &contractWatchers.ParticipantAddedEventWatcher{}) // Emit event: *contract.MpcManagerParticipantAdded

	// Subscribe events concerning local storage
	d.Subscribe(&events.GroupInfoStoredEvent{}, &contractWatchers.KeygenRequestAddedEventWatcher{}) // Emit event: *contract.MpcManagerKeygenRequestAdded
	d.Subscribe(&events.ParticipantInfoStoredEvent{}, nil)
	d.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, &contractWatchers.StakeRequestAddedEventWatcher{})   // Emit event: *contract.MpcManagerStakeRequestAdded
	d.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, &contractWatchers.StakeRequestStartedEventWatcher{}) // Emit event: *contract.MpcManagerStakeRequestStarted

	// Subscribe events from MpcManager
	d.Subscribe(&contract.MpcManagerParticipantAdded{}, &storageHandlers.ParticipantInfoStorer{}) // Emit event: *events.ParticipantInfoStoredEvent
	d.Subscribe(&contract.MpcManagerKeygenRequestAdded{}, nil)
	d.Subscribe(&contract.MpcManagerKeyGenerated{}, nil)
	d.Subscribe(&contract.MpcManagerStakeRequestAdded{}, nil)
	d.Subscribe(&contract.MpcManagerStakeRequestStarted{}, nil)

	return d
}
