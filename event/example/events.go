package main

import (
	"github.com/avalido/mpc-controller/event"
)

// my event types, registering to dispatcher
const (
	MyMessage event.EventType = iota
	MyWeather event.EventType = iota
)

// MessageEvent is an event type for messaging
type MessageEvent struct {
	Message string
}

// WeatherEvent is an event type for weather conditions
type WeatherEvent struct {
	Condition string
}
