# Softcom Jetstream EventStore Client - Go

#### How to Publish

```go
// Publish
package main

import (
	"encoding/json"
	"fmt"
	"github.com/softcomoss/jetstreamclient/options"
	jetstream "github.com/softcomoss/jetstreamclient/jsm"
	"log"
	"time"
)

func main() {
	ev, err := jetstream.Init(options.Options{
		ServiceName: "USERS",
		Address:     "localhost:4222",
	})
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	const topic = "test"

	data, err := json.Marshal(struct {
		FirstName string `json:"first_name"`
	}{
		FirstName: "Justice Nefe",
	})

	if err != nil {
		panic(err)
	}

	if err := ev.Publish(topic, data); err != nil {
		fmt.Print(err, " Error publishing.\n")
	}

	end := time.Now()

	diff := end.Sub(start)

	fmt.Printf("Start: %s, End: %s, Diff: %s", start, end, diff)
}

```

### How to Subscribe

```go
// Subscribe
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/softcomoss/jetstreamclient"
	jetstream "github.com/softcomoss/jetstreamclient/jsm"
	"github.com/softcomoss/jetstreamclient/options"

	"log"
	"time"
)

func main() {

	name := flag.String("name", "", "help message for flagname")
	flag.Parse()

	ev, err := jetstream.Init(options.Options{
		ServiceName: *name,
		Address:     "localhost:4222",
	})

	if err != nil {
		log.Fatal(err)
	}

	handleSubEv := func() error {
		const topic = "test"
		return ev.Subscribe(topic, func(event jetstreamclient.Event) {
			defer event.Ack()
			var pl struct {
				FirstName string `json:"first_name"`
			}
			
			if err := json.Unmarshal(event.Data(), &pl); err != nil {
				fmt.Print(err, " Err parsing event into pl.")
				return
			}

		}, options.NewSubscriptionOptions().SetSubscriptionType(options.Shared))
	}

	// You can use context.WithCancel()
	ev.Run(context.Background(), handleSubEv)
}

```
