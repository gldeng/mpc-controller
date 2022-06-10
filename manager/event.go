package manager

import (
	"time"
)

type Event struct {
	// Creator who create this event
	Creator string `json:"creator"`

	// ID is the unique database ID of the event.
	ID string `json:"id"`

	// Queue is the name of the queue. It defaults to the empty queue "".
	Queue string `json:"queue"`

	// Type maps event to a EventDispatcher.
	Type string `json:"type"`

	// Data is the data emitted with this event
	Data interface{} `json:"-"`

	// DataBytes must be the bytes of a valid JSON string
	DataBytes []byte `json:"dataBytes,omitempty"`

	// created when on event creation
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
