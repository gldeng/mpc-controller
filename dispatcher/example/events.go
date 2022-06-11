package main

import (
	"github.com/avalido/mpc-controller/dispatcher"
)

// my event types, registering to dispatcher
const (
	MyMessage dispatcher.EventType = iota
	MyWeather dispatcher.EventType = iota
)

// MessageEvent is an event type for messaging
type MessageEvent struct {
	Message string
}

// WeatherEvent is an event type for weather conditions
type WeatherEvent struct {
	Condition string
}
