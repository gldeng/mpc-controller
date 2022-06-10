package manager

import (
	"time"
)

// TaskPriority is the wrapper type for Task.Priority
type TaskPriority int16

// Some shortcut values for TaskPriority that can be any, but chances are high that one of these will be the most used.
const (
	TaskPriorityHighest TaskPriority = -32768
	TaskPriorityHigh    TaskPriority = -16384
	TaskPriorityDefault TaskPriority = 0
	TaskPriorityLow     TaskPriority = 16384
	TaskPriorityLowest  TaskPriority = 32767
)

type Task struct {
	// Creator who creates this task
	Creator string `json:"creator"`

	// ID is the unique database ID of the Task.
	ID string `json:"id"`

	// Queue is the name of the queue. It defaults to the empty queue "".
	Queue string `json:"queue"`

	// Priority is the priority of the task. The default priority is 0, and a
	// lower number means a higher priority.
	//
	// The highest priority is TaskPriorityHighest, the lowest one is TaskPriorityLowest
	Priority TaskPriority `json:"priority"`

	// RunAt is the time that this task should be executed. It defaults to now(),
	// meaning the task will execute immediately. Set it to a value in the future
	// to delay a task's execution.
	RunAt time.Time `json:"runAt"`

	// Type maps task to a worker func.
	Type string `json:"type"`

	// Args is the input data for this task
	Args interface{} `json:"-"`

	// ArgBytes must be the bytes of a valid JSON string
	ArgBytes []byte `json:"argBytes,omitempty"`

	// ReplyCh replies the result of the task
	// Returns is effective only when Error is nil.
	ReplyCh chan struct {
		Executor string
		Returns  interface{}
		Error    error
	} `json:"-"`

	// ReturnBytes is bytes of returned non-error value of the task
	ReturnBytes []byte `json:"returnBytes,omitempty"`

	// TaskCh is for delegating a task, which will be eventually executed by WorkFunc
	TaskCh chan *Task `json:"-"`

	// EventCh is for emitting an event, which will be immediately processed by EventDispatcher
	EventCh chan *Event `json:"-"`

	// ErrorCount is the number of times this task has attempted to run, but failed with an error.
	// It is ignored on task creation.
	ErrorCount int32 `json:"errorCount"`

	// LastError is the error message or stack trace from the last time the task failed. It is ignored on task creation.
	LastError error `json:"lastError"`

	// created when on task creation
	CreatedAt time.Time `json:"createdAt"`

	// updated when error occur during task execution
	ErrorAt time.Time `json:"errorAt"`

	// finished when task execution success
	FinishedAt time.Time `json:"finishedAt"`

	// IsCanceled denotes whether the task has been cancelled
	IsCanceled bool `json:"isCanceled"`
}
