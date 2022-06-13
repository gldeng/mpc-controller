package main

import (
	"context"
	contractWatchers "github.com/avalido/mpc-controller/contract/watchers"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/queue"
)

func Dispatcher(ctx context.Context, logger logger.Logger) *dispatcher.Dispatcher {
	// Create New dispatcher
	d := dispatcher.NewDispatcher(ctx, logger, queue.NewArrayQueue(1024), 1024)

	// Subscribe events concerning local storage
	d.Subscribe(&events.GroupInfoStoredEvent{}, &contractWatchers.KeygenRequestAddedEventWatcher{})
	d.Subscribe(&events.ParticipantInfoStoredEvent{}, nil)
	d.Subscribe(&events.GeneratedPubKeyInfoStoredEvent{}, nil)

	return d
}
