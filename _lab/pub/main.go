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
		ServiceName:         "sdfdk",
		Address:             "localhost:4222",
		AuthenticationToken: "TSdfsdf34o9432ksdkf24525209jc0vvnfn2349cc",
	})
	if err != nil {
		log.Fatal(err)
	}

	const topic = "loko"

	data, err := json.Marshal(struct {
		FirstName string `json:"first_name"`
	}{
		FirstName: "Justice Nefe",
	})

	if err != nil {
		panic(err)
	}

	for {
		if err := ev.Publish(topic, data); err != nil {
			fmt.Print(err, " Error publishing.\n")
		}
		time.Sleep(2 * time.Second)
	}
}
