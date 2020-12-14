package gonyexpress_test

import (
	"time"

	ge "github.com/SebastiaanPasterkamp/gonyexpress"
	"github.com/SebastiaanPasterkamp/gonyexpress/broker"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"strings"
	"testing"
)

func TestNewProducerConnectCloseWithError(t *testing.T) {
	p := ge.NewProducer("foo", "")

	if _, err := p.Connect(); !strings.Contains(err.Error(), "AMQP scheme must be either") {
		t.Errorf("Expected RabbitMQ URI Scheme err, but got '%+v'", err)
	}

	p.Close()
}

func TestProducer(t *testing.T) {
	p := ge.NewProducer("mock://", "test")
	m := p.Broker.(*broker.MockBroker)

	if _, err := p.Connect(); err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	defer p.Close()

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

	p.SendMessage(orig)

	msg, err := m.TakeMessage(1 * time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if msg == nil {
		t.Errorf("Expected a message, got nil")
	}

	if msg.TraceID != orig.TraceID {
		t.Errorf("TraceID changed. Have %q, want %q.", msg.TraceID, orig.TraceID)
	}

	d, ok := msg.Documents["input"]
	if !ok {
		t.Errorf("Reply should contain 'input' document")
	} else if string(d.Data) != "Hello" {
		t.Errorf("Unexpected 'input' content. Want %q, have %q", "Hello", d.Data)
	}
}
