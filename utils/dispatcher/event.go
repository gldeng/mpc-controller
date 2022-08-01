package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/stretchr/objx"
	"sort"
	"sync/atomic"
)

var eventNo uint64

func NewEvtNo() uint64 {
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
		EventNo: NewEvtNo(),
		EventID: misc.NewID(),

		Event: evt,
		Map:   mp,
	}
	return &evtObj
}

type EventObjects []*EventObject

func (s EventObjects) Len() int {
	return len(s)
}

func (s EventObjects) Less(i, j int) bool {
	return s[i].EventNo < s[j].EventNo
}

func (s EventObjects) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s EventObjects) Sort() {
	sort.Sort(s)
}
