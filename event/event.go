package event

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// Event can take all arguments of an event
type Event interface{}

// EventType is event type for an event
type EventType int

// EventObject contains event ID, event type, event creator, event created time, event as well context.
// Especially, context can convey extra necessary information, e.g. deadline, canceling, error and even k-v values.
type EventObject struct {
	EventID   uuid.UUID
	EventType EventType
	CreatedBy string
	CreatedAt time.Time

	Event   Event
	Context context.Context
}

// EventHandler is handler for events and takes any arguments
type EventHandler interface {
	Do(evtObj *EventObject)
}

// register event types, this is just a format, don't register here,
// rather use your own types like in the example/events.go

const (
	NoobEvent EventType = iota
)
