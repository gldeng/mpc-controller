// dispatcher is a lightweight in-memory event-driven framework for subscribing, publishing and dispatching events.

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

var (
	eventQueue Queue
	eventMap   map[string][]EventHandler
	eventLog   logger.Logger
	once       = &sync.Once{}
	mu         = &sync.Mutex{}
)

type Queue interface {
	Enqueue(e interface{})
	Dequeue() interface{}
	Empty() bool
	Full() bool
	Count() int
}

func init() {
	eventMap = make(map[string][]EventHandler)
}

// Subscribe to an event handler with any event.
// Please pass the pointer of empty exported event struct as event, e.g. &MySpecialEvent{}.
// Because underlying Subscribe use reflect to extract the type name of passed event as event type.
// In this way users do not need to define extra event type using enum data type,
// but must keep event type definition, or event schema as stable as possible,
// or any change to event schema could cause damage to data consistency.
func Subscribe(eH EventHandler, eT Event) {
	et := reflect.TypeOf(eT).String()

	if len(eventMap[et]) == 0 {
		eventMap[et] = make([]EventHandler, 0)
	}
	eventMap[et] = append(eventMap[et], eH)
}

// Publish can also receive event object and enqueue it.
// It is a package leve helper function to publish event without using a go channel.
func Publish(ctx context.Context, evtObj *EventObject) {
	enqueue(ctx, evtObj)
}

// doPublish concurrently run event handlers to the same event type.
// It waits until all the event handlers have returned.
// Note: event handlers may haven't finished their jobs yet,
// because they may run concurrently within their own gorutines.
// But this part is out of the dispatcher's control.
func doPublish(evtObj *EventObject) {
	et := reflect.TypeOf(evtObj.Event).String()

	wg := &sync.WaitGroup{}
	for _, eH := range eventMap[et] {
		wg.Add(1)
		eH := eH
		go func() {
			defer wg.Done()
			eH.Do(evtObj)
		}()
	}
	wg.Wait()
}

func enqueue(ctx context.Context, evtObj *EventObject) {
	mu.Lock()
	defer mu.Unlock()

	err := backoff.RetryFnExponentialForever(eventLog, ctx, func() error {
		if !eventQueue.Full() {
			eventQueue.Enqueue(evtObj)
			return nil
		}
		eventLog.Warn("The event queue is full!", []logger.Field{{"length", eventQueue.Count()}}...)
		return errors.New("The event queue is full!")
	})

	if err == nil {
		et := reflect.TypeOf(evtObj.Event).String()

		eventLog.Info("Event received and enqueued", []logger.Field{
			{"eventID", evtObj.EventID},
			{"eventType", et},
			{"event", evtObj.Event},
			{"createdBy", evtObj.CreatedBy},
			{"createdAt", evtObj.CreatedAt}}...)
	}
}

// run is a goroutine for receiving, enqueueing events.
// It also regularly dequeues and publishes events every one second.
func run(ctx context.Context, publisher chan *EventObject) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-publisher:
			enqueue(ctx, evtObj)
		case <-time.Tick(time.Second):
			if !eventQueue.Empty() {
				if evtObj, ok := eventQueue.Dequeue().(*EventObject); ok {
					doPublish(evtObj)
				}
			}
		}
	}
}

// NewEventPublisher makes a new publisher channel for events,
// which will run a goroutine for receiving and publishing events.
// Note: it's more safe to use publisher channel than call Publish() directly to publish events,
// when the length of buffered channel is smaller than the queue length limit value.
// This is because channel is concurrently safe but too many enqueuing operation may cause queue to panic
// especially when the queue is full. (this problem has been fixed within enqueue function by RetryFnExponentialForever)
// But when the channel is full the sender will get blocked, which may cause the whole program stop still
// if blocking problem does not be dealt properly.
func NewEventPublisher(ctx context.Context, logger logger.Logger, q Queue, bufLen int) chan *EventObject {
	once.Do(func() {
		eventLog = logger
		eventQueue = q
	})

	publisher := make(chan *EventObject, bufLen) // todo: deal with no-buffered channel block
	go run(ctx, publisher)
	return publisher
}
