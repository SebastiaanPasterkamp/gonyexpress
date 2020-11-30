package broker

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"strings"

	"github.com/streadway/amqp"
)

// Broker is an interface defining the bare functionality of a RabbitMQ
// connection.
type Broker interface {
	Connect() (<-chan amqp.Delivery, error)
	Close()
	SendMessage(msg payload.Message) error
}

// New creates either a RabbitMQ instance (default), or a MockBroker instance,
// ready to connect. Use a 'mock://' URI format for the MockBroker. A RabbitMQ
// URI typically starts with 'amqp://' or 'amqps://'.
func New(URI, qname string) Broker {
	if strings.HasPrefix(URI, "mock://") {
		return NewMockBroker()
	}
	return NewRabbitMQ(URI, qname)
}
