package main

import (
	ge "github.com/SebastiaanPasterkamp/gonyexpress"
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"flag"
	"fmt"
	"log"
)

func main() {
	log.Print("Go RabbitMQ Tutorial")

	rmq := flag.String(
		"rabbitmq", "amqp://guest:guest@127.0.0.1:5672/",
		"URL for RabbitMQ, e.g. 'amqp://user:pwd@host:port/'",
	)
	total := flag.Int(
		"total", 1, "Number of messages to send1.",
	)
	flag.Parse()

	p := ge.NewProducer(*rmq, "")

	_, err := p.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	for i := 0; i < *total; i++ {
		msg := payload.NewMessage(
			payload.Routing{
				Name:     "demo",
				Position: 0,
				Slip: []payload.Step{
					{
						Queue: "foo",
						Arguments: payload.Arguments{
							"duration": "1s",
						},
					},
					{
						Queue: "bar",
						Arguments: payload.Arguments{
							"duration": "2s",
						},
					},
				},
			},
			payload.MetaData{
				"origin": "producer",
			},
			payload.Documents{
				"input": payload.NewDocument(
					fmt.Sprintf("Hello world msg %d!", i),
					"text/plain",
					"",
				),
			},
		)

		err := p.SendMessage(msg)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s: Sent %+v\n", msg.TraceID, msg)
	}

	log.Println("Successfully Published Message(s) to Queue")
}
