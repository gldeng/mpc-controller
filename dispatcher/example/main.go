package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/queue"
	"github.com/google/uuid"
	"time"
)

func main() {
	// Create dispatcher
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	logger.DevMode = true
	log := logger.Default()
	d := dispatcher.NewDispatcher(ctx, log, queue.NewArrayQueue(1024), 1024)

	// Subscribe events to event handlers
	d.Subscribe(&MessageShower{d}, &MessageEvent{})
	d.Subscribe(&WeatherShower{}, &WeatherEvent{})

	// Publish events by Dispatcher channel.
	d.Channel() <- dispatcher.NewEventObject("MainFunction", &MessageEvent{Message: "Hello World"}, ctx)

	// Publish events by Dispatcher.Publish() method, in another gorutine
	go func() {
		myUuid, _ := uuid.NewUUID()
		d.Publish(ctx, &dispatcher.EventObject{
			EventID:   myUuid,
			CreatedBy: "MainFunction",
			CreatedAt: time.Now(),
			Event:     &MessageEvent{Message: "Nice to meet you!"},
			Context:   context.WithValue(ctx, "requestId", "fakeRequestId"),
		})
	}()

	<-ctx.Done()
}

type Publisher interface {
	Publish(ctx context.Context, evtObj *dispatcher.EventObject)
}

// MessageShower prints the message
type MessageShower struct {
	Publisher
}

func (m *MessageShower) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*MessageEvent); ok {
		fmt.Printf("Start taking action for MessageEvent %q from %q\n", evtObj.EventID, evtObj.CreatedBy)

		val := evtObj.Context.Value("requestId")
		if valStr, ok := val.(string); ok {
			fmt.Printf("RequestID: %q", valStr)
		}

		m.showMessage(evt)

		m.publishWeatherEvent(evtObj.Context, "Sunny")
	}
}

func (m *MessageShower) showMessage(evt *MessageEvent) {
	fmt.Println(evt.Message)
}

// Event handler can also publish event within its scope.
func (m *MessageShower) publishWeatherEvent(ctx context.Context, condition string) {
	uuid, _ := uuid.NewUUID()

	weatherEvtObj := &dispatcher.EventObject{
		EventID:   uuid,
		CreatedBy: "MessageShower",
		CreatedAt: time.Now(),

		Event: &WeatherEvent{
			Condition: condition,
		},
		Context: context.Background(),
	}

	m.Publish(ctx, weatherEvtObj)
}

// WeatherShower prints the weather condition
type WeatherShower struct {
}

func (m *WeatherShower) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*WeatherEvent); ok {
		fmt.Printf("Start taking action for WeatherEvent -%q from %q\n", evtObj.EventID, evtObj.CreatedBy)
		m.showWeather(evt)
	}
}

func (m *WeatherShower) showWeather(evt *WeatherEvent) {
	fmt.Println(evt.Condition)
}
