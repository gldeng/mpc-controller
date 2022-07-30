package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/misc"
	"time"
)

func main() {
	// Create dispatcher
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	logger.DevMode = true
	log := logger.Default()
	d := dispatcher.NewDispatcher(ctx, log, 1024)

	// Subscribe events to event handlers
	d.Subscribe(&MessageEvent{}, &MessageShower{d})
	d.Subscribe(&WeatherEvent{}, &WeatherShower{})

	// Publish events by Dispatcher.Publish() method, in another gorutine
	go func() {
		d.Publish(ctx, &dispatcher.EventObject{
			EventNo: dispatcher.AddEventCount(),
			EventID: misc.NewID(),
			Event:   &WeatherEvent{Condition: "Cloudy"},
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

func (m *MessageShower) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*MessageEvent); ok {
		m.showMessage(evt)

		m.publishWeatherEvent(ctx, "Sunny")
	}
}

func (m *MessageShower) showMessage(evt *MessageEvent) {
	fmt.Printf("\tMessage Content: %v\n", evt.Message)
}

// Event handler can also publish event within its scope.
func (m *MessageShower) publishWeatherEvent(ctx context.Context, condition string) {
	weatherEvtObj := &dispatcher.EventObject{
		EventNo: dispatcher.AddEventCount(),
		EventID: misc.NewID(),

		Event: &WeatherEvent{
			Condition: condition,
		},
	}

	m.Publish(ctx, weatherEvtObj)
}

// WeatherShower prints the weather condition
type WeatherShower struct {
}

func (m *WeatherShower) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*WeatherEvent); ok {
		m.showWeather(evt)
	}
}

func (m *WeatherShower) showWeather(evt *WeatherEvent) {
	fmt.Printf("\tWeather Condition: %v\n", evt.Condition)
}
