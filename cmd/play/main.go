package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"os"
	"regexp"
	"webiu/radio/station"
)

func main() {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	ctx := context.Background()

	rx := regexp.MustCompile("https://(.+.youtube.com|youtu.be)/(watch\\?v=(\\w+)|(\\w+))")
	_ = rx

	channelID := "C0UQ8TKLJ"

	bookmarks, err := api.ListBookmarks(channelID)

	if err != nil {
		panic(err)
	}

	for i := range bookmarks {
		bookmark := bookmarks[i]
		fmt.Printf("Title: %s\nId: %s\n", bookmark.Title, bookmark.AppID)
	}

	resp, err := api.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
	})

	for _, m := range resp.Messages {

		id, err := station.ExtractYoutubeID(m.Text)
		if err != nil {
			if errors.Is(err, station.ErrNotYoutubeLink) {
				continue
			}
			panic(err)
		}
		fmt.Printf("https://www.youtube.com/watch?v=%s\n", id)
	}
}
