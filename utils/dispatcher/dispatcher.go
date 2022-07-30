package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/misc"
	"reflect"
	"sync"
	"time"
)

type Dispatcher interface {
	Subscriber
	Publisher
}

type Subscriber interface {
	Subscribe(eT Event, eHs ...EventHandler)
}

type Publisher interface {
	Publish(ctx context.Context, evtObj *EventObject)
}

// dispatcher is a lightweight in-memory event-driven framework,
// dedicated for subscribing, publishing and dispatching events.
type dispatcher struct {
	id          string
	eventLogger logger.Logger
	eventChan   chan *EventObject
	eventMap    map[string][]EventHandler
	subscribeMu *sync.Mutex
}

// NewDispatcher makes a new dispatcher for users to subscribe events,
// and runs a goroutine for receiving and publishing event objects.
func NewDispatcher(ctx context.Context, logger logger.Logger, name string, evtChanLen int) Dispatcher {
	dispatcher := &dispatcher{
		id:          name + "_dispatcher_" + misc.NewID()[:4],
		eventLogger: logger,
		eventChan:   make(chan *EventObject, evtChanLen),
		eventMap:    make(map[string][]EventHandler),
		subscribeMu: new(sync.Mutex),
	}
	go dispatcher.run(ctx)
	return dispatcher
}

// Subscribe to event handler(s) with any event.
// Please pass the pointer of empty exported event struct as event, e.g. &MySpecialEvent{}.
func (d *dispatcher) Subscribe(eT Event, eHs ...EventHandler) {
	d.subscribeMu.Lock()
	defer d.subscribeMu.Unlock()

	if eT != nil && eHs != nil {
		et := reflect.TypeOf(eT).String()
		for _, eH := range eHs {
			if eH != nil {
				d.eventMap[et] = append(d.eventMap[et], eH)
				eh := reflect.TypeOf(eH).String()
				d.eventLogger.Info(d.id+" subscribed an event", []logger.Field{
					{"event", et},
					{"handler", eh}}...)
			}
		}
	}
}

// Publish sends the received event object to underlying channel.
func (d *dispatcher) Publish(ctx context.Context, evtObj *EventObject) {
	et := reflect.TypeOf(evtObj.Event).String()
	ehs := d.eventMap[et]
	if len(ehs) == 0 {
		d.eventLogger.Warn(d.id+" no subscriber", []logger.Field{{"eventType", et}}...)
		return
	}
	d.eventChan <- evtObj
}

// run is a goroutine for receiving and publishing events.
func (d *dispatcher) run(ctx context.Context) {
	t := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			d.eventLogger.Debug(d.id+" dispatcher health stats",
				[]logger.Field{{"cachedEvents", len(d.eventChan)}}...)
			if float64(len(d.eventChan)) > float64(cap(d.eventChan))*0.8 {
				d.eventLogger.Warn(d.id+" dispatcher cached too many events",
					[]logger.Field{{"cachedEvents", len(d.eventChan)},
						{"cacheCapacity", cap(d.eventChan)}}...)
			}
		case evtObj := <-d.eventChan:
			et := reflect.TypeOf(evtObj.Event).String()
			ehs := d.eventMap[et]
			if len(ehs) > 0 {
				for _, eh := range ehs {
					eh.Do(ctx, evtObj)
				}
			}
		}
	}
}
