package payload_test

import (
	"testing"

	"github.com/SebastiaanPasterkamp/gonyexpress/payload"
)

func TestNewMessage(t *testing.T) {
	msg := payload.NewMessage(
		payload.Routing{
			Name:     "from_byte_slice",
			Position: 1,
			Slip: []payload.Step{
				{Queue: "step-1"},
				{
					Queue: "step-2",
					Arguments: payload.Arguments{
						"foo": "bar",
					},
					ErrorHandling: payload.ErrorHandling{
						MaxRetries: 3,
						Rewind:     1,
					},
				},
			},
		},
		payload.MetaData{
			"meta": "data",
		},
		payload.Documents{
			"doc": payload.NewDocument("test", "text/plain", ""),
		},
	)

	if len(msg.TraceID) != 36 {
		t.Errorf("Unexpected TraceID format. Expected length of %d, but got %d.",
			36, len(msg.TraceID))
	}

	if msg.Routing.Name != "from_byte_slice" {
		t.Errorf("Unexpected Routing.Name. Have %q, want %q.",
			msg.Routing.Name, "from_byte_slice")
	}
	if msg.Routing.Position != 1 {
		t.Errorf("Unexpected Routing.Position. Have %d, want %d.",
			msg.Routing.Position, 1)
	}

	step, err := msg.CurrentStep()
	if err != nil {
		t.Errorf("Unexpected error: %+v.", err)
	}

	foo, ok := step.Arguments["foo"]
	if !ok {
		t.Errorf("Missing step argument 'foo' in %+v.", step.Arguments)
	} else if foo.(string) != "bar" {
		t.Errorf("Unexpected value for 'foo' Step Argument. Have %q, want %q.",
			foo, "bar")
	}

	meta, ok := msg.MetaData["meta"]
	if !ok {
		t.Errorf("Missing MetaData 'meta' in %+v.", msg.MetaData)
	} else if meta.(string) != "data" {
		t.Errorf("Unexpected value for 'meta' MetaData. Have %q, want %q.",
			meta, "data")
	}

	doc, ok := msg.Documents["doc"]
	if !ok {
		t.Errorf("Missing Documents 'doc' in %+v.", msg.Documents)
	} else if string(doc.Data) != "test" {
		t.Errorf("Unexpected value for 'doc' Data in %+v. Have %q, want %q.",
			doc, doc.Data, "test")
	}
}

func TestNewMessageForRoute(t *testing.T) {
	msg := payload.NewMessageForRoute(
		"testing",
		payload.MetaData{
			"meta": "data",
		},
		payload.Documents{
			"doc": payload.NewDocument("test", "text/plain", ""),
		},
	)

	if len(msg.TraceID) != 36 {
		t.Errorf("Unexpected TraceID format. Expected length of %d, but got %d.",
			36, len(msg.TraceID))
	}

	if msg.Routing.Name != "testing" {
		t.Errorf("Unexpected Routing.Name. Have %q, want %q.",
			msg.Routing.Name, "testing")
	}
	if msg.Routing.Position != 0 {
		t.Errorf("Unexpected Routing.Position. Have %d, want %d.",
			msg.Routing.Position, 0)
	}

	step, err := msg.CurrentStep()
	if err != nil {
		t.Errorf("Unexpected error: %+v.", err)
	}
	if step.Queue != "post-office" {
		t.Errorf("Unexpected Step.Queue. Have %q, want %q.",
			step.Queue, "post-office")
	}

	meta, ok := msg.MetaData["meta"]
	if !ok {
		t.Errorf("Missing MetaData 'meta' in %+v.", msg.MetaData)
	} else if meta.(string) != "data" {
		t.Errorf("Unexpected value for 'meta' MetaData. Have %q, want %q.",
			meta, "data")
	}

	doc, ok := msg.Documents["doc"]
	if !ok {
		t.Errorf("Missing Documents 'doc' in %+v.", msg.Documents)
	} else if string(doc.Data) != "test" {
		t.Errorf("Unexpected value for 'doc' Data in %+v. Have %q, want %q.",
			doc, doc.Data, "test")
	}
}

func TestMessageFromByteSlice(t *testing.T) {
	msg, err := payload.MessageFromByteSlice([]byte(`{"broken"`))
	if err == nil {
		t.Errorf("Expected malformed JSON error")
	}

	msg, err = payload.MessageFromByteSlice([]byte(`{
		"routing": {
			"name": "from_byte_slice",
			"position": 1,
			"slip": [
				{
					"queue": "step-1"
				},
				{
					"queue": "step-2",
					"arguments": {
						"foo": "bar"
					},
					"on_error": {
						"max_retries": 3,
						"rewind": 1
					}
				}
			]
		},
		"trace_id": "f00-b4r",
		"metadata": {
			"meta": "data"
		},
		"documents": {
			"doc": {
				"content_type": "text/plain",
				"data": "test"
			}
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}

	if msg.TraceID != "f00-b4r" {
		t.Errorf("Unexpected TraceID. Have %q, want %q.",
			msg.TraceID, "f00-b4r")
	}

	if msg.Routing.Name != "from_byte_slice" {
		t.Errorf("Unexpected Routing.Name. Have %q, want %q.",
			msg.Routing.Name, "from_byte_slice")
	}
	if msg.Routing.Position != 1 {
		t.Errorf("Unexpected Routing.Position. Have %d, want %d.",
			msg.Routing.Position, 1)
	}

	step, err := msg.CurrentStep()
	if err != nil {
		t.Errorf("Unexpected error: %+v.", err)
	}

	foo, ok := step.Arguments["foo"]
	if !ok {
		t.Errorf("Missing step argument 'foo' in %+v.", step.Arguments)
	} else if foo.(string) != "bar" {
		t.Errorf("Unexpected value for 'foo' Step Argument. Have %q, want %q.",
			foo, "bar")
	}

	meta, ok := msg.MetaData["meta"]
	if !ok {
		t.Errorf("Missing MetaData 'meta' in %+v.", msg.MetaData)
	} else if meta.(string) != "data" {
		t.Errorf("Unexpected value for 'meta' MetaData. Have %q, want %q.",
			meta, "data")
	}

	doc, ok := msg.Documents["doc"]
	if !ok {
		t.Errorf("Missing Documents 'doc' in %+v.", msg.Documents)
	} else if string(doc.Data) != "test" {
		t.Errorf("Unexpected value for 'doc' Data in %+v. Have %q, want %q.",
			doc, doc.Data, "test")
	}
}

func TestCurrentStepBad(t *testing.T) {
	msg := payload.NewMessage(
		payload.Routing{
			Name:     "current_step",
			Position: 1,
			Slip: []payload.Step{
				{Queue: "post-office"},
			},
		},
		payload.MetaData{},
		payload.Documents{},
	)

	step, err := msg.CurrentStep()
	if err == nil {
		t.Errorf("Expected error, got %+v, nil instead.", step)
	}
}

func TestAdvance(t *testing.T) {
	msg := payload.NewMessage(
		payload.Routing{
			Name:     "test-advance",
			Position: -1,
			Slip: []payload.Step{
				{Queue: "here"},
				{Queue: "there"},
			},
		},
		payload.MetaData{
			"meta": "data",
		},
		payload.Documents{
			"doc": payload.NewDocument("test", "text/plain", ""),
		},
	)

	next, err := msg.Advance(nil, nil)
	if err == nil {
		t.Errorf("Expected error, got nil.")
	}
	if next != nil {
		t.Errorf("Unexpected next message: %+v", next)
	}

	msg.Routing.Position = 1
	next, err = msg.Advance(nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if next != nil {
		t.Errorf("Unexpected next message: %+v", next)
	}

	msg.Routing.Position = 0
	next, err = msg.Advance(
		&payload.Documents{
			"new": payload.NewDocument("success", "text/plain", ""),
		},
		&payload.MetaData{
			"add": "more",
		},
	)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if next == nil {
		t.Errorf("Expected next message, got nil")
	} else {

		if next.TraceID != msg.TraceID {
			t.Errorf("TraceID changed. Have %q, want %q.", next.TraceID, msg.TraceID)
		}

		if next.Routing.Position != 1 {
			t.Errorf("Expected position advance. Have %d, want %d.",
				next.Routing.Position, 1)
		}

		d, ok := next.Documents["doc"]
		if !ok {
			t.Errorf("Next should contain 'doc' document")
		} else if string(d.Data) != "test" {
			t.Errorf("Unexpected 'doc' content. Want %q, have %q", "test", d.Data)
		}

		d, ok = next.Documents["new"]
		if !ok {
			t.Errorf("Next should contain 'new' document")
		} else if string(d.Data) != "success" {
			t.Errorf("Unexpected 'new' content. Want %q, have %q", "success", d.Data)
		}

		m, ok := next.MetaData["meta"]
		if !ok {
			t.Errorf("Next should contain 'meta' metadata")
		} else if m.(string) != "data" {
			t.Errorf("Unexpected 'meta' metadata. Want %q, have %+v", "data", m)
		}

		m, ok = next.MetaData["add"]
		if !ok {
			t.Errorf("Next should contain 'add' metadata")
		} else if m.(string) != "more" {
			t.Errorf("Unexpected 'add' metadata. Want %q, have %+v", "more", m)
		}
	}
}

func TestRetry(t *testing.T) {
	msg := payload.NewMessage(
		payload.Routing{
			Name:     "test-advance",
			Position: 1,
			Slip: []payload.Step{
				{Queue: "back"},
				{Queue: "again"},
			},
		},
		payload.MetaData{
			"meta": "data",
		},
		payload.Documents{
			"doc": payload.NewDocument("test", "text/plain", ""),
		},
	)

	// retry not configured
	retry, err := msg.Retry()
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if retry != nil {
		t.Errorf("Unexpected retry message: %+v", retry)
	}

	// retry configured, but invalid rewind
	msg.Routing.Slip[1].MaxRetries = 3
	msg.Routing.Slip[1].Rewind = 2
	retry, err = msg.Retry()
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if retry != nil {
		t.Errorf("Unexpected retry message: %+v", retry)
	}

	// maxed out attempts
	msg.Routing.Slip[1].MaxRetries = 3
	msg.Routing.Slip[1].Attempt = 3
	retry, err = msg.Retry()
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if retry != nil {
		t.Errorf("Unexpected retry message: %+v", retry)
	}

	// valid settings
	msg.Routing.Slip[1].Attempt = 0
	msg.Routing.Slip[1].Rewind = 1
	retry, err = msg.Retry()
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if retry == nil {
		t.Errorf("Expected next message, got nil")
	} else {
		if retry.TraceID != msg.TraceID {
			t.Errorf("TraceID changed. Have %q, want %q.", retry.TraceID, msg.TraceID)
		}

		if retry.Routing.Position != 0 {
			t.Errorf("Expected position rewind. Have %d, want %d.",
				retry.Routing.Position, 0)
		}

		d, ok := retry.Documents["doc"]
		if !ok {
			t.Errorf("Retry should contain 'doc' document")
		} else if string(d.Data) != "test" {
			t.Errorf("Unexpected 'doc' content. Want %q, have %q", "test", d.Data)
		}

		m, ok := retry.MetaData["meta"]
		if !ok {
			t.Errorf("Retry should contain 'meta' metadata")
		} else if m.(string) != "data" {
			t.Errorf("Unexpected 'meta' metadata. Want %q, have %+v", "data", m)
		}
	}
}
