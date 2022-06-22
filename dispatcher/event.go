package dispatcher

import (
	"context"
	"reflect"
)

// Event can take all arguments of an event
type Event interface{}

// EventObject contains event ID, event type, event creator, event created time, event as well context.
// Especially, context can convey extra necessary information, e.g. deadline, canceling, error and even k-v values.
// ParentEvent is the event that trigger the event handler to emit the current event.
// If there's no parent(root event), ParentEvtNo and ParentEvtID should be the default value,
// Every event in the same event stream should share the same EvtStreamNo and EvtStreamID value,
// which originally generated by the root event.
type EventObject struct {
	RootEvtType string
	RootEvtID   string
	RootEvtNo   uint64

	EvtStreamNo uint64 // starting from 1
	EvtStreamID string

	ParentEvtNo uint64
	ParentEvtID string

	EventStep int // offset in an event stream, starting from 1

	EventNo   uint64 // starting from 1
	EventID   string
	CreatedBy string
	CreatedAt string // timestamp

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

// NewRootEventObject is convenience to create an root EventObject.
func NewRootEventObject(createdBy string, evt Event, ctx context.Context) *EventObject {
	et := reflect.TypeOf(evt).String()

	rootEvtID := NewID()
	rootEvtNo := AddEventCount()

	evtObj := EventObject{
		RootEvtType: et,
		RootEvtID:   rootEvtID,
		RootEvtNo:   rootEvtNo,

		EvtStreamNo: AddEventStreamCount(),
		EvtStreamID: NewID(),

		ParentEvtNo: uint64(0),
		ParentEvtID: "",

		EventStep: 1,

		EventNo:   rootEvtNo,
		EventID:   rootEvtID,
		CreatedBy: createdBy,
		CreatedAt: NewTimestamp(),

		Event:   evt,
		Context: ctx,
	}
	return &evtObj
}

// NewEventObjectFromParent is convenience to create child EventObject from given parent EventObject.
func NewEventObjectFromParent(parentEvtObj *EventObject, createdBy string, evt Event, ctx context.Context) *EventObject {
	childEvtObj := EventObject{
		RootEvtType: parentEvtObj.RootEvtType,
		RootEvtID:   parentEvtObj.RootEvtID,
		RootEvtNo:   parentEvtObj.RootEvtNo,

		EvtStreamNo: parentEvtObj.EvtStreamNo,
		EvtStreamID: parentEvtObj.EvtStreamID,

		ParentEvtNo: parentEvtObj.ParentEvtNo,
		ParentEvtID: parentEvtObj.ParentEvtID,

		EventStep: parentEvtObj.EventStep + 1,

		EventNo:   AddEventCount(),
		EventID:   NewID(),
		CreatedBy: createdBy,
		CreatedAt: NewTimestamp(),

		Event:   evt,
		Context: ctx,
	}
	return &childEvtObj
}

// NewEventObject is convenience to create an EventObject.
func NewEventObject(rootEvtType string, rootEvtID string, rootEvtNo uint64,
	evtStreamNo uint64, evtStreamID string,
	parentEvtNo uint64, parentEvtID string, parentEvtStep int,
	createdBy string, evt Event, ctx context.Context) *EventObject {
	evtObj := EventObject{
		RootEvtType: rootEvtType,
		RootEvtID:   rootEvtID,
		RootEvtNo:   rootEvtNo,

		EvtStreamNo: evtStreamNo,
		EvtStreamID: evtStreamID,

		ParentEvtNo: parentEvtNo,
		ParentEvtID: parentEvtID,

		EventStep: parentEvtStep + 1,

		EventNo:   AddEventCount(),
		EventID:   NewID(),
		CreatedBy: createdBy,
		CreatedAt: NewTimestamp(),

		Event:   evt,
		Context: ctx,
	}
	return &evtObj
}
