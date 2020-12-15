package broker

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"encoding/json"

	"github.com/streadway/amqp"
)

// AMQPReplyTo is a "magic" direct reply-to channel a Producer can use to
// request progress or result messages.
const AMQPReplyTo = "amq.rabbitmq.reply-to"

// RabbitMQ is a RabbitMQ Broker implementation which can be used by consumers,
// and producers alike.
type RabbitMQ struct {
	// URI of the RabbitMQ instance
	URI string
	// Name of the RabbitMQ Queue to subscribe to
	qname string
	// conn is the active connection to the RabbitMQ service
	conn *amqp.Connection
	// ch is the RabbitMQ channel by which Messages are sent
	ch *amqp.Channel
}

// NewRabbitMQ creates a RabbitMQ instance ready to connect.
func NewRabbitMQ(URI, qname string) *RabbitMQ {
	return &RabbitMQ{
		URI:   URI,
		qname: qname,
	}
}

// Connect opens up a RabbitMQ connection and returns a channel through which
// Messages are delivered.
func (r *RabbitMQ) Connect(prefetch int) (<-chan amqp.Delivery, error) {
	var err error
	r.conn, err = amqp.Dial(r.URI)
	if err != nil {
		return nil, err
	}

	r.ch, err = r.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = r.ch.Qos(
		prefetch, // prefetch count
		0,        // prefetch size
		false,    // global
	)
	if err != nil {
		return nil, err
	}

	if r.qname == "" {
		// Producer-only mode
		return nil, nil
	}

	// TODO: Queue definitions should happen outside
	if r.qname != AMQPReplyTo {
		_, err = r.ch.QueueDeclare(
			r.qname, // name
			true,    // durable
			false,   // delete when unused
			false,   // exclusive
			false,   // no-wait
			nil,     // arguments for plugins
		)
		if err != nil {
			return nil, err
		}
	}

	return r.ch.Consume(
		r.qname, // key
		"",      // consumer
		false,   // auto ack
		false,   // exclusive
		false,   // no local
		false,   // no wait
		nil,     // args for plugins
	)
}

// Close terminates the RabbitMQ channel and connection. Should be used when
// running a Producer, after Connect is called. Automatically called after
// Shutdown for a running Consumer.
func (r *RabbitMQ) Close() {
	if r.ch != nil {
		r.ch.Close()
		r.ch = nil
	}
	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}
}

// SendMessage sends a message onto the message's current Slip queue
func (r *RabbitMQ) SendMessage(msg payload.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.ch.Publish(
		"", // exchange
		msg.Routing.Slip[msg.Routing.Position].Queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			CorrelationId: msg.TraceID,
			ContentType:   "application/json",
			Body:          body,
		},
	)
}
