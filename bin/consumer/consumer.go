package main

import (
	ge "github.com/SebastiaanPasterkamp/gonyexpress"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	log.Print("Go RabbitMQ Tutorial")

	rmq := flag.String(
		"rabbitmq", "amqp://guest:guest@127.0.0.1:5672/",
		"URL for RabbitMQ, e.g. 'amqp://user:pwd@host:port/'",
	)
	qname := flag.String(
		"queue", "TestQueue", "Name of the AMQP queue to write to.",
	)
	flag.Parse()

	forever := make(chan bool)

	c := ge.NewConsumer(*rmq, *qname, 4, operation)
	err := c.Run()
	if err != nil {
		log.Printf("Failed to launch component: %+v\n", err)
		close(forever)
		c.Shutdown()
	}
	defer c.Shutdown()

	<-forever
}

func operation(
	traceID string, md payload.MetaData, args payload.Arguments, docs payload.Documents,
) (*payload.Documents, *payload.MetaData, error) {
	log.Printf("%s - Recieved Message: %+v\n", traceID, docs)

	var err error
	sleep := 0 * time.Second
	if duration, ok := args["duration"]; ok {
		sleep, err = time.ParseDuration(duration.(string))
		if err != nil {
			log.Printf("%s - Failed to parse duration '%+s': %+v\n", traceID, duration, err)
			return nil, nil, fmt.Errorf("Failed to parse duration: '%+s': %+v", duration, err)
		}
	}
	time.Sleep(sleep)

	docid := "output"
	if d, ok := args["docid"]; ok {
		docid = d.(string)
	}

	pl := payload.Documents{
		docid: payload.NewDocument(traceID, "text/plain", ""),
	}

	return &pl, nil, nil
}
