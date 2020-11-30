package gonyexpress

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/broker"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"fmt"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// NewConsumer creates a Consumer Component instance ready to connect to the
// rabbitmq + queue and execute the operator function for every recieved
// message.
func NewConsumer(URI, qname string, workers int, operator Operator) Component {
	return Component{
		Broker:   broker.New(URI, qname),
		operator: operator,
		workers:  workers,
		wg:       sync.WaitGroup{},
	}
}

// Operator is a function to be executed for every message received by the
// Consumer. The returned payload is advanced to the next step on the routing
// slip.
type Operator func(
	traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
) (*payload.Documents, *payload.MetaData, error)

// Run launches the Component as a background service.
func (c *Component) Run() error {
	if c.operator == nil {
		return fmt.Errorf("cannot Run without operator")
	}
	if c.workers < 1 {
		return fmt.Errorf("cannot Run without workers")
	}

	msgs, err := c.Connect()
	if err != nil {
		c.Close()
		return errors.Wrap(err, "Failed to connect to RabbitMQ")
	}
	log.Info("Successfully Connected to our RabbitMQ Instance")

	c.shutdown = make(chan bool)

	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(msgs)
	}

	log.Info("Component running")
	return nil
}

// IsShuttingDown returns a channel to be closed when the Consumer is shutting
// down.
func (c *Component) IsShuttingDown() <-chan bool {
	return c.shutdown
}

func (c *Component) worker(msgs <-chan amqp.Delivery) {
	defer c.wg.Done()

	log.Info("Launched worker...")

	for {
		select {
		case <-c.IsShuttingDown():
			log.Warning("Shutting down worker...")
			return

		case d := <-msgs:
			msg, err := payload.MessageFromByteSlice(d.Body)

			if err != nil {
				log.Errorf("%s - Bad message: %+v in %+v\n",
					d.CorrelationId, err, d.Body)
				// TODO: enable retry
				d.Nack(false, false)
				continue
			}
			step, err := msg.CurrentStep()
			if err != nil {
				log.Errorf("%s - Bad message: %+v in %+v\n",
					d.CorrelationId, err, d.Body)
				// TODO: enable retry
				d.Nack(false, false)
				continue
			}

			pl, md, err := c.operator(
				msg.TraceID,
				msg.MetaData,
				step.Arguments,
				msg.Documents,
			)

			var next *payload.Message
			if err != nil {
				log.Errorf("%s - Failed to process message: %+v\n",
					d.CorrelationId, err)
				next, err = msg.Retry()
				if err != nil {
					log.Errorf("%s - Failed to produc retry message: %+v\n",
						d.CorrelationId, err)
				}
			} else {
				next, err = msg.Advance(pl, md)
				if err != nil {
					log.Errorf("%s - Failed to produc next message: %+v\n",
						d.CorrelationId, err)
				}
			}

			if next != nil {
				err = c.Broker.SendMessage(*next)
				if err != nil {
					log.Errorf("%s - Failed to send message: %+v\n",
						d.CorrelationId, err)
					d.Nack(false, false)
					continue
				}
			}

			d.Ack(false)
		}
	}
}

// Shutdown will notify all workers to stop, and wait for all to finish.
func (c *Component) Shutdown() {
	log.Println("Shutting down")
	if c.shutdown == nil {
		return
	}

	select {
	case <-c.shutdown:
		// Not running
	default:
		close(c.shutdown)
		c.wg.Wait()
		c.Close()
		c.shutdown = nil
	}
}
