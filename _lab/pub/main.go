package main

import (
	"encoding/json"
	"fmt"
	"github.com/softcomoss/jetstreamclient/options"

	jetstream "github.com/softcomoss/jetstreamclient/jsm"
	"log"
	"sync"
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

	wg := &sync.WaitGroup{}
	for i := 1; i < 20000; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if err := ev.Publish(topic, data); err != nil {
				fmt.Print(err, " Error publishing.\n")
			}
		}(wg)
	}
	wg.Wait()
	end := time.Now()

	diff := end.Sub(start)

	fmt.Printf("Start: %s, End: %s, Diff: %s", start, end, diff)
}
