package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/duythinht/shout/station"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	ctx := context.Background()

	channelID := "C0UQ8TKLJ"

	s := station.New(token, channelID)

	queue, err := s.Watch(ctx)

	if err != nil {
		panic(err)
	}

	for {
		if queue.Size() > 0 {
			link := queue.Poll()
			fmt.Println(link)
		}

		time.Sleep(1 * time.Second)
	}
}
