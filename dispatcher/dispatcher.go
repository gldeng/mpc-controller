package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"reflect"
	"sync"
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

type Channeller interface {
	Channel() chan *EventObject
}

// DispatcherImpl is a lightweight in-memory event-driven framework,
// dedicated for subscribing, publishing and dispatching events.
type DispatcherImpl struct {
	eventLogger logger.Logger
	eventChan   chan *EventObject
	eventMap    map[string][]EventHandler
	subscribeMu *sync.Mutex
}

// NewDispatcher makes a new dispatcher for users to subscribe events,
// and runs a goroutine for receiving and publishing event objects.
func NewDispatcher(ctx context.Context, logger logger.Logger, evtChanLen int) *DispatcherImpl {
	dispatcher := &DispatcherImpl{
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
// Because underlying Subscribe use reflect to extract the type name of passed event as event type.
// In this way users do not need to define extra event type using enum data type,
// but must keep event type definition, or event schema as stable as possible,
// or any change to event schema could cause damage to data consistency.
func (d *DispatcherImpl) Subscribe(eT Event, eHs ...EventHandler) {
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
func (d *DispatcherImpl) Publish(ctx context.Context, evtObj *EventObject) {
	//d.publish(ctx, evtObj)
	d.eventChan <- evtObj
}

// Channel exposes the underlying channel for users to send event objects externally,
// e.g. DispatcherImpl.Channel <- &myEventObject
func (d *DispatcherImpl) Channel() chan *EventObject {
	return d.eventChan
}

// run is a goroutine for receiving and publishing events.
func (d *DispatcherImpl) run(ctx context.Context) {
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
// Under the hood its use Workshop to do the amazing scheduling.
func (d *DispatcherImpl) publish(ctx context.Context, evtObj *EventObject) {
	et := reflect.TypeOf(evtObj.Event).String()
	ehs := d.eventMap[et]
	if len(ehs) > 0 {
		for _, eh := range ehs {
			go eh.Do(ctx, evtObj)
		}
	}
}
