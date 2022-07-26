package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/avalido/mpc-controller/utils/work"
	"time"
)

func main() {
	// Create dispatcher
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	logger.DevMode = true
	log := logger.Default()
	ws := work.NewWorkshop(log, time.Second*300, 1024, make(chan *work.Task, 1024))
	d := dispatcher.NewDispatcher(ctx, log, 1024, ws)

	// Subscribe events to event handlers
	d.Subscribe(&MessageEvent{}, &MessageShower{d})
	d.Subscribe(&WeatherEvent{}, &WeatherShower{})

	// Publish events by Dispatcher channel.
	d.Channel() <- dispatcher.NewRootEventObject("MainFunction", &MessageEvent{Message: "Hello World"}, ctx)

	// Publish events by Dispatcher.Publish() method, in another gorutine
	go func() {
		d.Publish(ctx, &dispatcher.EventObject{
			EvtStreamNo: dispatcher.AddEventStreamCount(),
			EvtStreamID: misc.NewID(),
			ParentEvtNo: uint64(0),
			ParentEvtID: "",
			EventStep:   1,
			EventNo:     dispatcher.AddEventCount(),
			EventID:     misc.NewID(),
			CreatedBy:   "MainFunction",
			CreatedAt:   misc.NewTimestamp(),
			Event:       &WeatherEvent{Condition: "Cloudy"},
			Context:     ctx,
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

		m.publishWeatherEvent(evtObj.EvtStreamNo, evtObj.EvtStreamID, evtObj.EventNo, evtObj.EventID, evtObj.EventStep, evtObj.Context, "Sunny")
	}
}

func (m *MessageShower) showMessage(evt *MessageEvent) {
	fmt.Printf("\tMessage Content: %v\n", evt.Message)
}

// Event handler can also publish event within its scope.
func (m *MessageShower) publishWeatherEvent(evtStreamNo uint64, evtStreamID string, parentEvtNo uint64, parentEvtID string, parentEvtStep int, ctx context.Context, condition string) {
	weatherEvtObj := &dispatcher.EventObject{
		EvtStreamNo: evtStreamNo,
		EvtStreamID: evtStreamID,
		ParentEvtNo: parentEvtNo,
		ParentEvtID: parentEvtID,
		EventStep:   parentEvtStep + 1,
		EventNo:     dispatcher.AddEventCount(),
		EventID:     misc.NewID(),
		CreatedBy:   "MessageShower",
		CreatedAt:   misc.NewTimestamp(),

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

func (m *WeatherShower) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*WeatherEvent); ok {
		m.showWeather(evt)
	}
}

func (m *WeatherShower) showWeather(evt *WeatherEvent) {
	fmt.Printf("\tWeather Condition: %v\n", evt.Condition)
}
