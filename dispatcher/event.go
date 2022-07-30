package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/stretchr/objx"
)

type Event interface{}

type EventObject struct {
	EventNo uint64 // starting from 1
	EventID string
	Event   Event
	Map     objx.Map
}

type EventHandler interface {
	Do(ctx context.Context, evtObj *EventObject)
}

func NewEvtObj(evt Event, mp objx.Map) *EventObject {
	evtObj := EventObject{
		EventNo: AddEventCount(),
		EventID: misc.NewID(),

		Event: evt,
		Map:   mp,
	}
	return &evtObj
}
