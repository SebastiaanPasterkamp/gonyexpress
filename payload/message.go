package payload

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Message is a simple example message
type Message struct {
	Routing   `json:"routing"`
	TraceID   string `json:"trace_id"`
	MetaData  `json:"metadata,omitempty"`
	Documents `json:"documents,omitempty"`
}

// Routing contains the routing name, slip, and current step number.
type Routing struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
	Slip     []Step `json:"slip"`
}

// Step is a single processing step in a routing slip
type Step struct {
	Queue         string `json:"queue"`
	Arguments     `json:"arguments,omitempty"`
	ErrorHandling `json:"on_error,omitempty"`
}

// Arguments is a set of key-value pairs containing arguments specific to a Step
type Arguments map[string]interface{}

// ErrorHandling declares how failed steps are to be handled. Currently focusses
// on retry limits, and (partially) rewinding the route position.
type ErrorHandling struct {
	MaxRetries int `json:"max_retries"`
	Attempt    int `json:"attempt,omitempty"`
	Rewind     int `json:"rewind,omitempty"`
}

// MetaData is a set of key-value pairs containing general information about a Message
type MetaData map[string]interface{}

// MessageFromByteSlice unmarshals a JSON byte slice into a Message.
func MessageFromByteSlice(b []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(b, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewMessage creates a new Message with a unique TraceID, and attaches the
// routing slip, metadatam and documents.
func NewMessage(route Routing, meta MetaData, docs Documents) Message {
	traceID, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("could not create UUID: %v", err)
	}

	return Message{
		TraceID:   traceID.String(),
		Routing:   route,
		MetaData:  meta,
		Documents: docs,
	}
}

// NewMessageForRoute creates a new Message with a unique TraceID to send to the
// post-office for the route by name.
func NewMessageForRoute(route string, meta MetaData, docs Documents) Message {
	traceID, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("could not create UUID: %v", err)
	}

	return Message{
		TraceID: traceID.String(),
		Routing: Routing{
			Name: route,
			Slip: []Step{
				{Queue: "post-office"},
			},
		},
		MetaData:  meta,
		Documents: docs,
	}
}

// CurrentStep return the Step at the current position in the routing slip.
func (msg Message) CurrentStep() (*Step, error) {
	if msg.Routing.Position < 0 || msg.Routing.Position >= len(msg.Routing.Slip) {
		return nil, fmt.Errorf("Invalid Position: %d / %d",
			msg.Routing.Position, len(msg.Routing.Slip))
	}
	return &msg.Routing.Slip[msg.Routing.Position], nil
}

// Advance creates a new Message based on the current message, but with updated
// documents and advanced routing position. Returns nil if there is no next step.
func (msg Message) Advance(pl *Documents, md *MetaData) (*Message, error) {
	if msg.Routing.Position < 0 {
		return nil, fmt.Errorf("dropping invalid message at step %d / %d",
			msg.Routing.Position+1, len(msg.Routing.Slip))
	}
	if msg.Routing.Position+1 >= len(msg.Routing.Slip) {
		log.Printf("At step %d / %d. Finished route...", msg.Routing.Position+1, len(msg.Routing.Slip))
		return nil, nil
	}

	log.Printf("Advancing to step %d / %d. Still enroute...", msg.Routing.Position+2, len(msg.Routing.Slip))

	return &Message{
		Routing: Routing{
			Name:     msg.Routing.Name,
			Position: msg.Routing.Position + 1,
			Slip:     msg.Slip,
		},
		TraceID:   msg.TraceID,
		MetaData:  msg.combineMetaData(md),
		Documents: msg.combineDocuments(pl),
	}, nil
}

// Retry creates a new Message based on the current message, but with updated
// attempt count, and possibly a partially reset Position. Returns nil if the
// number of retries has been exhausted.
func (msg Message) Retry() (*Message, error) {
	step, err := msg.CurrentStep()
	if err != nil {
		return nil, err
	}

	if step.Attempt >= step.MaxRetries {
		log.Printf("At attempt %d / %d. Giving up...", step.Attempt+1, step.MaxRetries)
		return nil, nil
	}

	if step.Rewind < 0 {
		log.Printf("Rewind must be positive. Dropping invalid route...")
		return nil, nil
	}

	if msg.Routing.Position-step.Rewind < 0 {
		log.Printf("Rewinding beyond begining. Dropping invalid route...")
		return nil, nil
	}

	log.Printf("Retrying step %d / %d. Still enroute...",
		msg.Routing.Position-step.Rewind+1, len(msg.Routing.Slip))

	step.Attempt++

	return &Message{
		Routing: Routing{
			Name:     msg.Routing.Name,
			Position: msg.Routing.Position - step.Rewind,
			Slip:     msg.Slip,
		},
		TraceID:   msg.TraceID,
		MetaData:  msg.MetaData,
		Documents: msg.Documents,
	}, nil
}

func (msg Message) combineMetaData(update *MetaData) MetaData {
	if update == nil {
		return msg.MetaData
	}

	m := MetaData{}
	for k, v := range msg.MetaData {
		m[k] = v
	}
	for k, v := range *update {
		m[k] = v
	}
	return m
}

func (msg Message) combineDocuments(update *Documents) Documents {
	if update == nil {
		return msg.Documents
	}

	p := Documents{}
	for k, v := range msg.Documents {
		p[k] = v
	}
	for k, v := range *update {
		p[k] = v
	}
	return p
}
