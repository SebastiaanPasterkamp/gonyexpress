package broker

import (
	"time"

	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"encoding/json"

	"github.com/streadway/amqp"
)

// MockBroker is a Mock Broker implementation which can be used in tests.
type MockBroker struct {
	inc chan amqp.Delivery
	out chan amqp.Delivery
}

// NewMockBroker creates a Mock Broker instance ready for testing.
func NewMockBroker() MockBroker {
	return MockBroker{
		inc: make(chan amqp.Delivery, 10),
		out: make(chan amqp.Delivery, 10),
	}
}

// Connect only serves to complete the Broker interface. It returns the mock
// message channel
func (m MockBroker) Connect() (<-chan amqp.Delivery, error) {
	return m.inc, nil
}

// Close closes the test queue.
func (m MockBroker) Close() {
	close(m.inc)
	close(m.out)
}

// SendMessage sends a message onto the outgoing queue
func (m MockBroker) SendMessage(msg payload.Message) error {
	m.addMessageToQueue(m.out, msg)
	return nil
}

// DeliverMessage puts a message onto the incoming queue
func (m MockBroker) DeliverMessage(msg payload.Message) {
	m.addMessageToQueue(m.inc, msg)
}

func (m MockBroker) addMessageToQueue(q chan amqp.Delivery, msg payload.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	q <- amqp.Delivery{
		// Properties
		ContentType:   "application/json",
		CorrelationId: msg.TraceID,
		DeliveryMode:  amqp.Persistent,
		Body:          body,
	}

	return nil
}

// TakeMessage pops a message from the outgoing queue, of one is available. Does
// not block.
func (m MockBroker) TakeMessage(d time.Duration) (*payload.Message, error) {
	select {
	case d := <-m.out:
		return payload.MessageFromByteSlice(d.Body)
	case <-time.After(d):
		return nil, nil
	}
}
