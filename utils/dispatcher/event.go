package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/stretchr/objx"
	"sync/atomic"
)

var eventNo uint64

func newEvtNo() uint64 {
	return atomic.AddUint64(&eventNo, 1)
}

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
		EventNo: newEvtNo(),
		EventID: misc.NewID(),

		Event: evt,
		Map:   mp,
	}
	return &evtObj
}
