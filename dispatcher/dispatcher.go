package dispatcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"sync"
	"time"
)

var (
	eventQueue Queue
	eventMap   map[EventType][]EventHandler
	eventLog   logger.Logger
	once       = &sync.Once{}
)

type Queue interface {
	Enqueue(e interface{})
	Dequeue() interface{}
	Empty() bool
}

func init() {
	eventMap = make(map[EventType][]EventHandler)
}

// Subscribe to an event handler
func Subscribe(eH EventHandler, eT EventType) {
	if len(eventMap[eT]) == 0 {
		eventMap[eT] = make([]EventHandler, 0)
	}
	eventMap[eT] = append(eventMap[eT], eH)
}

// Publish can also receive event object and enqueue it.
// It is a package leve helper function to publish event without using a go channel.
func Publish(evtObj *EventObject) {
	enqueue(evtObj)
}

// doPublish concurrently run event handlers to the same event type.
// It waits until all the event handlers have finished their jobs.
func doPublish(evtObj *EventObject) {
	wg := &sync.WaitGroup{}
	for _, eH := range eventMap[evtObj.EventType] {
		wg.Add(1)
		eH := eH
		go func() {
			defer wg.Done()
			eH.Do(evtObj)
		}()
	}
	wg.Wait()
}

func enqueue(evtObj *EventObject) {
	eventQueue.Enqueue(evtObj)
	eventLog.Info("Event received and enqueued", []logger.Field{
		{"eventID", evtObj.EventID},
		{"eventType", evtObj.EventType},
		{"event", evtObj.Event},
		{"createdBy", evtObj.CreatedBy},
		{"createdAt", evtObj.CreatedAt}}...)
}

// run is a goroutine for receiving, enqueueing events.
// It also regularly dequeues and publishes events every one second.
func run(ctx context.Context, publisher chan *EventObject) {
	for {
		select {
		case <-ctx.Done():
			return
		case evtObj := <-publisher:
			enqueue(evtObj)
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
// especially when the queue is full.
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
