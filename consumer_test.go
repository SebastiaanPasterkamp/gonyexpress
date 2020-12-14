package gonyexpress_test

import (
	"time"

	ge "github.com/SebastiaanPasterkamp/gonyexpress"
	"github.com/SebastiaanPasterkamp/gonyexpress/broker"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"fmt"
	"strings"
	"testing"
)

func TestNewRunShutdownWithError(t *testing.T) {
	operator := func(
		traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
	) (*payload.Documents, *payload.MetaData, error) {
		t.Error("Operator should not be called")
		return nil, nil, fmt.Errorf("Should not be called")
	}
	c := ge.NewConsumer("foo", "bar", 1, operator)

	select {
	case <-c.IsShuttingDown():
		t.Error("Should not be called")
	default:
		// noop
	}

	if err := c.Run(); !strings.Contains(err.Error(), "AMQP scheme must be either") {
		t.Errorf("Expected RabbitMQ URI Scheme err, but got '%+v'", err)
	}

	c.Shutdown()
}

func TestConsumer(t *testing.T) {
	operator := func(
		traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
	) (*payload.Documents, *payload.MetaData, error) {
		d := &payload.Documents{
			"test": payload.NewDocument("passed", "text/plain", ""),
		}
		return d, nil, nil
	}

	c := ge.NewConsumer("mock://", "test", 1, operator)
	m := c.Broker.(*broker.MockBroker)

	if err := c.Run(); err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}

	orig := payload.NewMessage(
		payload.Routing{
			Name: "test-consumer",
			Slip: []payload.Step{
				{Queue: "foo"},
				{Queue: "bar"},
			},
		},
		payload.MetaData{},
		payload.Documents{
			"input": payload.NewDocument(
				"Hello",
				"text/plain",
				"",
			),
		},
	)

	m.DeliverMessage(orig)

	msg, err := m.TakeMessage(1 * time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if msg == nil {
		t.Errorf("Expected a message, got nil")
	} else {
		if msg.TraceID != orig.TraceID {
			t.Errorf("TraceID changed. Have %q, want %q.", msg.TraceID, orig.TraceID)
		}

		d, ok := msg.Documents["input"]
		if !ok {
			t.Errorf("Reply should contain 'input' document")
		} else if string(d.Data) != "Hello" {
			t.Errorf("Unexpected 'input' content. Want %q, have %q", "Hello", d.Data)
		}

		d, ok = msg.Documents["test"]
		if !ok {
			t.Errorf("Reply should contain 'test' document")
		} else if string(d.Data) != "passed" {
			t.Errorf("Unexpected 'test' content. Want %q, have %q", "passed", d.Data)
		}
	}

	c.Shutdown()
}

func TestConsumerPing(t *testing.T) {
	operator := func(
		_ string, _ payload.MetaData, _ payload.Arguments, _ payload.Documents,
	) (*payload.Documents, *payload.MetaData, error) {
		t.Fatal("Not expected to be called")
		return nil, nil, nil
	}

	c := ge.NewConsumer("mock://", "test", 1, operator)
	m := c.Broker.(*broker.MockBroker)

	if err := c.Run(); err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}

	orig := payload.NewMessage(
		payload.Routing{
			Name: "test-consumer",
			Slip: []payload.Step{
				{Queue: "foo"},
				{Queue: "bar"},
			},
		},
		payload.MetaData{"ping": true},
		payload.Documents{},
	)

	m.DeliverMessage(orig)

	msg, err := m.TakeMessage(1 * time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if msg == nil {
		t.Errorf("Expected a message, got nil")
	} else {
		if msg.TraceID != orig.TraceID {
			t.Errorf("TraceID changed. Have %q, want %q.", msg.TraceID, orig.TraceID)
		}
	}

	c.Shutdown()
}

func TestBadInit(t *testing.T) {
	operator := func(
		traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
	) (*payload.Documents, *payload.MetaData, error) {
		t.Error("Operator should not be called")
		return nil, nil, fmt.Errorf("Should not be called")
	}

	c := ge.NewConsumer("mock://", "test", 0, operator)
	p := ge.NewProducer("mock://", "test")

	err := c.Run()
	if err == nil {
		t.Errorf("Expected error, received nil")
	} else if !strings.Contains(err.Error(), "without worker") {
		t.Errorf("Expected error. Have %+v, want something with 'without workers'.", err)
	}
	defer c.Shutdown()

	err = p.Run()
	if err == nil {
		t.Errorf("Expected error, received nil")
	} else if !strings.Contains(err.Error(), "without operator") {
		t.Errorf("Expected error. Have %+v, want something with 'without operator'.", err)
	}
	defer p.Shutdown()
}

func TestConsumerFailure(t *testing.T) {
	operator := func(
		traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
	) (*payload.Documents, *payload.MetaData, error) {
		d := &payload.Documents{
			"test": payload.NewDocument("failed", "text/plain", ""),
		}
		return d, nil, fmt.Errorf("Please fail this message")
	}

	c := ge.NewConsumer("mock://", "test", 1, operator)
	m := c.Broker.(*broker.MockBroker)

	if err := c.Run(); err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}

	orig := payload.NewMessageForRoute(
		"just-testing",
		payload.MetaData{},
		payload.Documents{
			"input": payload.NewDocument(
				"Hello",
				"text/plain",
				"",
			),
		},
	)

	m.DeliverMessage(orig)

	msg, err := m.TakeMessage(1 * time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if msg != nil {
		t.Errorf("Unexpected message %+v", msg)
	}

	c.Shutdown()
}
