package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duythinht/shout/station"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")

	channelID := "C0UQ8TKLJ"

	s := station.New(token, channelID)

	stop, err := s.Welcome(context.Background())

	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stop()
		os.Exit(1)
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
