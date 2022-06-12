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
	// Create publisher
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*11)
	defer cancel()
	logger.DevMode = true
	log := logger.Default()
	publisher := dispatcher.NewEventPublisher(ctx, log, queue.NewArrayQueue(1024), 1024)

	// Subscribe events to event handlers
	dispatcher.Subscribe(&MessageShower{publisher}, &MessageEvent{})
	dispatcher.Subscribe(&WeatherShower{}, &WeatherEvent{})

	// Publish events.
	// Events can also be published in event handler,
	// by calling event.Publish() or write to publisher channel.
	myUuid, _ := uuid.NewUUID()
	publisher <- &dispatcher.EventObject{
		EventID:   myUuid,
		CreatedBy: "MainFunction",
		CreatedAt: time.Now(),
		Event:     &MessageEvent{Message: "Hello World"},
		Context:   ctx,
	}

	myUuid, _ = uuid.NewUUID()
	publisher <- &dispatcher.EventObject{
		EventID:   myUuid,
		CreatedBy: "MainFunction",
		CreatedAt: time.Now(),
		Event:     &MessageEvent{Message: "Nice to meet you!"},
		Context:   context.WithValue(ctx, "requestId", "fakeRequestId"),
	}

	<-ctx.Done()
}

// MessageShower prints the message
type MessageShower struct {
	publisher chan *dispatcher.EventObject
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
		m.publishWeatherEvent(evtObj.Context, "Rainy")

		go func() {
			m.publishWeatherEvent(evtObj.Context, "Sunny")
		}()

		go func() {
			m.publishWeatherEvent(evtObj.Context, "Rainy")
		}()
	}
}

func (m *MessageShower) showMessage(evt *MessageEvent) {
	fmt.Println(evt.Message)
}

// Event handler can also publish event within its scope.
// This is just a demonstration.
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

	// There are two ways to publish an event
	//m.publisher <- weatherEvtObj // todo: deal with no-buffered channel block
	dispatcher.Publish(ctx, weatherEvtObj) // this way also works.
}

// WeatherShower prints the weather condition
type WeatherShower struct {
}

func (m *WeatherShower) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*WeatherEvent); ok {
		fmt.Printf("Start taking action for WeatherEvent -%q from %q\n", evtObj.EventID, evtObj.CreatedBy)
		m.ShowWeather(evt)
	}
}

func (m *WeatherShower) ShowWeather(evt *WeatherEvent) {
	fmt.Println(evt.Condition)
}
