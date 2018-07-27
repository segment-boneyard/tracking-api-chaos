package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/analytics-go"
)

func main() {
	endpoint := "https://api.segment.build"
	if e := os.Getenv("ENDPOINT"); e != "" {
		endpoint = e
	}
	fmt.Println("Sending events to", endpoint)
	client, err := analytics.NewWithConfig(os.Getenv("WRITE_KEY"), analytics.Config{
		Endpoint: endpoint,
		Verbose:  true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	for range time.Tick(1000 * time.Millisecond) {
		err := client.Enqueue(analytics.Track{
			Event:  "Trace Tested",
			UserId: "Collin",
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(time.Now())
	}
}
