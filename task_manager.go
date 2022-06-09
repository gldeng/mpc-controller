package mpc_controller

import (
	"context"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding task manager, which is fully event-driven

type Event interface{}

type Task interface {
	Do(ctx context.Context, input Event) (output Event, err error)
}

type TaskCreator interface {
	NewTask() Task
}

type TaskManager interface {
	Start(ctx context.Context) error
	Register(evt Event, task TaskCreator) error
}
