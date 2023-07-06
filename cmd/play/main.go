package main

import (
	"context"
	"fmt"
	"os"
	"webiu/radio/station"

	"github.com/slack-go/slack"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	ctx := context.Background()

	api := slack.New(token)
	channelID := "C0UQ8TKLJ"

	bookmarks, err := api.ListBookmarks(channelID)

	if err != nil {
		panic(err)
	}

	for i := range bookmarks {
		bookmark := bookmarks[i]
		fmt.Printf("Title: %s\nId: %s\n", bookmark.Title, bookmark.AppID)
	}

	s := station.New(token, channelID)
	play, err := s.History(ctx)
	if err != nil {
		panic(err)
	}
	for play.Size() > 0 {
		fmt.Printf("link: %s\n", play.Poll())
	}
}
