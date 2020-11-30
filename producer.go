package gonyexpress

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/broker"
)

// NewProducer creates a Component instance ready to Connect to the rabbitmq,
// but it does not launch any workers and doesn't automatically consume incoming
// messages. The queue can be empty if the Producer is not interested in the
// result or progress updates, otherwise it is recommended to set the queue to:
// 'amq.rabbitmq.reply-to' (AMQPReplyTo) for a Direct Reply-To.
func NewProducer(URI, qname string) Component {
	return Component{
		Broker:  broker.New(URI, qname),
		workers: 0,
	}
}
