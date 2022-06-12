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
	d.Channel() <- dispatcher.NewEventObject(uuid.UUID{}, "MainFunction", &MessageEvent{Message: "Hello World"}, ctx)

	// Publish events by Dispatcher.Publish() method, in another gorutine
	go func() {
		myUuid, _ := uuid.NewUUID()
		d.Publish(ctx, &dispatcher.EventObject{
			LastEvent: uuid.UUID{},
			EventID:   myUuid,
			CreatedBy: "MainFunction",
			CreatedAt: time.Now(),
			Event:     &WeatherEvent{Condition: "Cloudy"},
			Context:   ctx,
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
		m.showMessage(evt)

		m.publishWeatherEvent(evtObj.EventID, evtObj.Context, "Sunny")
	}
}

func (m *MessageShower) showMessage(evt *MessageEvent) {
	fmt.Printf("\tMessage Content: %v\n", evt.Message)
}

// Event handler can also publish event within its scope.
func (m *MessageShower) publishWeatherEvent(lastEvt uuid.UUID, ctx context.Context, condition string) {
	uuid, _ := uuid.NewUUID()

	weatherEvtObj := &dispatcher.EventObject{
		LastEvent: lastEvt,
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
		m.showWeather(evt)
	}
}

func (m *WeatherShower) showWeather(evt *WeatherEvent) {
	fmt.Printf("\tWeather Condition: %v\n", evt.Condition)
}
