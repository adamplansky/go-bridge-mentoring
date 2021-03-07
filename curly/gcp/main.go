package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	//rt := rounttripper.New()

	//c := http.Client{
	//	Transport: rt,
	//	Timeout:   10 * time.Second,
	//}
	//opt := option.WithHTTPClient(&c)

	client, err := pubsub.NewClient(ctx, "gdrive-adam-plansky")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	topic, err := client.CreateTopic(context.Background(), "topic1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(topic)

}
