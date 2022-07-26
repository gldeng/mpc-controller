package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/work"

	//"github.com/avalido/mpc-controller/utils/work"
	"reflect"
	"sync"
)

type Workshop interface {
	AddTask(ctx context.Context, t *work.Task)
}

type Dispatcherrer interface {
	Subscriber
	Publisher
	Channeller
}

type DispatcherClaasic interface {
	Subscriber
	Publisher
}

type Subscriber interface {
	Subscribe(eT Event, eHs ...EventHandler)
}

type Publisher interface {
	Publish(ctx context.Context, evtObj *EventObject)
}

type Channeller interface {
	Channel() chan *EventObject
}

// Dispatcher is a lightweight in-memory event-driven framework,
// dedicated for subscribing, publishing and dispatching events.
type Dispatcher struct {
	eventLogger logger.Logger
	eventChan   chan *EventObject
	eventMap    map[string][]EventHandler
	subscribeMu *sync.Mutex
	workshop    Workshop
}

// NewDispatcher makes a new dispatcher for users to subscribe events,
// and runs a goroutine for receiving and publishing event objects.
func NewDispatcher(ctx context.Context, logger logger.Logger, evtChanLen int, ws Workshop) *Dispatcher {
	dispatcher := &Dispatcher{
		eventLogger: logger,
		eventChan:   make(chan *EventObject, evtChanLen),
		eventMap:    make(map[string][]EventHandler),
		subscribeMu: new(sync.Mutex),
		workshop:    ws,
	}
	go dispatcher.run(ctx)
	return dispatcher
}

// Subscribe to event handler(s) with any event.
// Please pass the pointer of empty exported event struct as event, e.g. &MySpecialEvent{}.
// Because underlying Subscribe use reflect to extract the type name of passed event as event type.
// In this way users do not need to define extra event type using enum data type,
// but must keep event type definition, or event schema as stable as possible,
// or any change to event schema could cause damage to data consistency.
func (d *Dispatcher) Subscribe(eT Event, eHs ...EventHandler) {
	d.subscribeMu.Lock()
	defer d.subscribeMu.Unlock()

	if eT != nil && eHs != nil {
		et := reflect.TypeOf(eT).String()
		for _, eH := range eHs {
			if eH != nil {
				d.eventMap[et] = append(d.eventMap[et], eH)
				eh := reflect.TypeOf(eH).String()
				d.eventLogger.Info("Subscribed an event", []logger.Field{
					{"event", et},
					{"handler", eh}}...)
			}
		}
	}
}

// Publish sends the received event object to underlying channel.
// All event objects will be serialized in a queue and published later in FIFO order.
func (d *Dispatcher) Publish(ctx context.Context, evtObj *EventObject) {
	d.eventChan <- evtObj
}

// Channel exposes the underlying channel for users to send event objects externally,
// e.g. Dispatcher.Channel <- &myEventObject
func (d *Dispatcher) Channel() chan *EventObject {
	return d.eventChan
}

// run is a goroutine for receiving, enqueueing events.
func (d *Dispatcher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-d.eventChan:
			d.publish(ctx, evtObj)
		}
	}
}

// publish concurrently run event handlers to the same event type.
// It waits until all the event handlers have returned.
// Note: event handlers may haven't finished their jobs yet,
// because they may run concurrently within their own gorutines.
// But this part is out of the dispatcher's control.
func (d *Dispatcher) publish(ctx context.Context, evtObj *EventObject) {
	et := reflect.TypeOf(evtObj.Event).String()
	ehs := d.eventMap[et]
	if len(ehs) > 0 {
		task := work.Task{
			Args:    evtObj,
			Ctx:     ctx,
			WorkFns: workFnFromEventHandlers(ehs),
		}
		d.workshop.AddTask(ctx, &task)
	}
}
