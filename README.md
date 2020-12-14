# Gony Express

The `gonyexpress` is a Proof of Concept package to quickly create `golang` AMQP
enabled Components. This package implements the "Routing Slip Pattern", and does
all the heavy lifting, from:

* Receiving a Message
* Unpacking a Message
* `<your magic here>`
* Re-packaging the Message
* Advancing a Message / or Retrying a failed Message
* Acknowledging a Message has been handled

All a Gony Express `Component` needs is to do it's own bit of magic.

## Example

```golang
package main

import (
	ge "github.com/SebastiaanPasterkamp/gonyexpress"
	. "github.com/SebastiaanPasterkamp/gonyexpress/payload"
)

func main() {
	c := ge.NewConsumer(
        "amqp://guest:guest@127.0.0.1:5672/",
        "example", // incoming queue
        4,         // worker count
        magic,     // your magic
    )

    forever := make(chan bool)
    c.Run()
	defer c.Shutdown()
	<-forever
}

func magic(
	traceID string, md MetaData, args Arguments, docs Documents,
) (*Documents, *MetaData, error) {
    // Work with the incoming general metadata, your step specific arguments,
    // and the attached documents.

    // Attach more documents and/or inject more metadata to the next message
	out := Documents{
		"example": NewDocument("your result here", "text/plain", ""),
	}

    // Return
	return &out, nil, nil
}
```

# Future work

The Gony Express will work on extra features, such as `opentracing` support,
diagnostics, improved logging, configuration, and more.
