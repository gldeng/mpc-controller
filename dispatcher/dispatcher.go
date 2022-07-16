package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"time"
)

type Queue interface {
	Enqueue(e interface{})
	Dequeue() interface{}
	List() []interface{}
	Empty() bool
	Full() bool
	Count() int
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

	publishChan chan *EventObject
	eventQueue  Queue
	eventMap    map[string][]EventHandler

	once        *sync.Once
	queueMu     *sync.Mutex
	subscribeMu *sync.Mutex
}

// NewDispatcher makes a new dispatcher for users to subscribe events,
// and runs a goroutine for receiving and publishing event objects.
func NewDispatcher(ctx context.Context, logger logger.Logger, q Queue, bufLen int) *Dispatcher {
	dispatcher := &Dispatcher{
		eventLogger: logger,

		publishChan: make(chan *EventObject, bufLen),
		eventQueue:  q,
		eventMap:    make(map[string][]EventHandler),

		once:        new(sync.Once),
		queueMu:     new(sync.Mutex),
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
func (p *Dispatcher) Publish(ctx context.Context, evtObj *EventObject) {
	select {
	case <-ctx.Done():
		return
	case p.publishChan <- evtObj:
	}
}

// Channel exposes the underlying channel for users to send event objects externally,
// e.g. Dispatcher.Channel <- &myEventObject
func (d *Dispatcher) Channel() chan *EventObject {
	return d.publishChan
}

// run is a goroutine for receiving, enqueueing events.
// It also regularly dequeues and publishes events every 500 milliseconds.
func (d *Dispatcher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-d.publishChan:
			et := reflect.TypeOf(evtObj.Event).String()
			d.eventLogger.Info("Received an event", []logger.Field{
				{"eventType", et},
				{"eventNo", evtObj.EventNo},
				{"eventValue", evtObj.Event},

				//{"eventStep", evtObj.EventStep},

				//{"parentEvtNo", evtObj.ParentEvtNo},
				//{"parentEvtID", evtObj.ParentEvtID},

				//{"rootEvtType", evtObj.RootEvtType},
				//{"rootEvtID", evtObj.RootEvtID},
				//{"rootEvtNo", evtObj.RootEvtNo},
				//
				//{"evtStreamNo", evtObj.EvtStreamNo},
				//{"evtStreamID", evtObj.EvtStreamID},
				//
				//{"eventID", evtObj.EventID},
				//{"createdBy", evtObj.CreatedBy},
				//{"createdAt", evtObj.CreatedAt}}...
			}...,
			)
			if len(d.eventMap[et]) > 0 { // only enqueue when there exist(s) event handler
				d.enqueue(ctx, evtObj)
			}
		case <-time.Tick(time.Millisecond * 500):
			if !d.eventQueue.Empty() {
				if evtObj, ok := d.eventQueue.Dequeue().(*EventObject); ok {
					d.publish(ctx, evtObj)
				}
			}
		}
	}
}

// enqueue receives and serializes the given event object in queue,
// which will later be published  in FIFO order.
func (d *Dispatcher) enqueue(ctx context.Context, evtObj *EventObject) {
	d.queueMu.Lock()
	defer d.queueMu.Unlock()

	et := reflect.TypeOf(evtObj.Event).String()

	err := backoff.RetryFnExponentialForever(d.eventLogger, ctx, func() error {
		if !d.eventQueue.Full() {
			d.eventQueue.Enqueue(evtObj)
			return nil
		}
		d.eventLogger.Warn("The event queue is full!", []logger.Field{{"length", d.eventQueue.Count()}}...)
		return errors.New("The event queue is full!")
	})

	if err != nil {
		d.eventLogger.Error("Failed to enqueue an event", []logger.Field{
			{"error", err},
			{"eventType", et},
			{"eventValue", evtObj.Event}}...)
	}

	var evtStatMap = map[string]int{}
	var ets []string
	evtObjs := d.eventQueue.List()
	for _, evtObj := range evtObjs {
		et := reflect.TypeOf(evtObj.(*EventObject).Event).String()
		ets = append(ets, et)
		evtStatMap[et]++
	}
	d.eventLogger.Debug("Current events in queue", []logger.Field{
		{"totalCount", d.eventQueue.Count()},
		{"EventStats", evtStatMap},
		{"EventQueue", ets}}...)
}

// publish concurrently run event handlers to the same event type.
// It waits until all the event handlers have returned.
// Note: event handlers may haven't finished their jobs yet,
// because they may run concurrently within their own gorutines.
// But this part is out of the dispatcher's control.
func (d *Dispatcher) publish(ctx context.Context, evtObj *EventObject) {
	et := reflect.TypeOf(evtObj.Event).String()

	wg := &sync.WaitGroup{}
	for _, eH := range d.eventMap[et] {
		wg.Add(1)
		eH := eH
		go func() {
			defer wg.Done()
			eH.Do(ctx, evtObj)
		}()
	}
	wg.Wait()
}
