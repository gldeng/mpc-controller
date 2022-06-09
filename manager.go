package mpc_controller

import (
	"context"
	"net/http"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding Manager

// Event stands for something happened and usually trigger to task.
type Event interface{}
type Task func(ctx context.Context, evt Event)

// Request can be used to request external resources during task execution.
type Request interface{}
type Serve func(ctx context.Context, req Request)

// Manager is essentially an event-driven multiplexer or mediator
// which routes messages between external resources as supportive services and
// tasks that focus on dealing with business processes and scenarios
type Manager interface {
	Start(ctx context.Context) error

	RegisterEvent(task Task, evt ...Event)
	Event(ctx context.Context, evt Event) // emit an event by the caller

	RegisterRequest(srv Serve, req ...Request)
	Request(ctx context.Context, req Request) // issue a request by the caller
	http.ServeMux
}
