package dispatcher

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// Event can take all arguments of an event
type Event interface{}

// EventObject contains event ID, event type, event creator, event created time, event as well context.
// Especially, context can convey extra necessary information, e.g. deadline, canceling, error and even k-v values.
type EventObject struct {
	LastEvent uuid.UUID

	EventID   uuid.UUID
	CreatedBy string
	CreatedAt time.Time

	Event   Event
	Context context.Context
}

// EventHandler is handler for events and takes any arguments.
// An event handler can be stateless or stateful.
// It's better to take effective measures for state persistence and resume for a stateful event handler.
// Besides, event handler can return immediately and choose to execute task in separate gorutines.
// In this way keep in mind do not let gorutine leak. It's handler's responsibility to ensure this security.
type EventHandler interface {
	Do(evtObj *EventObject)
}

// NewEventObject is convenience to create an EventObject.
func NewEventObject(lastEvt uuid.UUID, createdBy string, evt Event, ctx context.Context) *EventObject {
	myUuid, _ := uuid.NewUUID()
	evtObj := EventObject{
		LastEvent: lastEvt,

		EventID:   myUuid,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),

		Event:   evt,
		Context: ctx,
	}
	return &evtObj
}
