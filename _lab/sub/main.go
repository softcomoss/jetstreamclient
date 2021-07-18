package main

import (
	"flag"
	"fmt"
	"github.com/softcomoss/jetstreamclient"
	jetstream "github.com/softcomoss/jetstreamclient/jsm"
	"github.com/softcomoss/jetstreamclient/options"

	"log"
)

func main() {

	name := flag.String("name", "", "help message for flagname")
	flag.Parse()

	fmt.Print(*name, " Name")

	ev, err := jetstream.Init(options.Options{
		//ContentType: "application/json",
		ServiceName:         *name,
		Address:             "localhost:4222",
		AuthenticationToken: "TSdfsdf34o9432ksdkf24525209jc0vvnfn2349cc",
		//Codecs: codec.De
	})

	if err != nil {
		log.Fatal(err)
	}
	err = ev.Subscribe("loko", func(event jetstreamclient.Event) {
		fmt.Print(string(event.Data()), " Event data")
		event.Ack()
	})

	fmt.Print(err, " Err Stream")

	//events.NewEventHandler(ev).Listen()
}
