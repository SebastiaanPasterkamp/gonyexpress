package gonyexpress

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/broker"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"sync"

	"github.com/streadway/amqp"
)

// Component is a RabbitMQ consumer / producer to do the heavy lifting for routing
type Component struct {
	// Broker is a utility wrapper around a RabbitMQ connection.
	Broker broker.Broker
	// Operator is thread-safe function called for every message
	operator Operator
	// Worker channel to communicate start shutdown
	shutdown chan bool
	// wg is the WaitGroup synchronizing the shutdown of all Workers
	wg sync.WaitGroup
	// Workers is the number of workers to spawn
	workers int
}

// Connect opens up a RabbitMQ connection and returns a channel through which
// Messages are delivered.
func (c *Component) Connect() (<-chan amqp.Delivery, error) {
	return c.Broker.Connect()
}

// Close terminates the RabbitMQ channel and connection. Should be used when
// running a Producer, after Connect is called. Automatically called after
// Shutdown for a running Consumer.
func (c *Component) Close() {
	c.Broker.Close()
}

// SendMessage sends a message onto the message's current Slip queue
func (c *Component) SendMessage(msg payload.Message) error {
	return c.Broker.SendMessage(msg)
}
