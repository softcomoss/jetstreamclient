package main

import (
	"encoding/json"
	"fmt"
	"github.com/softcomoss/jetstreamclient/_lab/sub/events/topics"
	"github.com/softcomoss/jetstreamclient/options"

	jetstream "github.com/softcomoss/jetstreamclient/jsm"
	"log"
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

	//const topic = "loko"

	data, err := json.Marshal(struct {
		FirstName string `json:"first_name"`
	}{
		FirstName: "Justice Nefe",
	})

	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
			fmt.Print("Firing event...\n")
			if err := ev.Publish(topics.KYCUserBVNVerificationUpdated, data); err != nil {
				fmt.Print(err, " Error publishing.\n")
			}
		//time.Sleep(1 * time.Second)
	}
}
