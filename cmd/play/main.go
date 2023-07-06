package main

import (
	"context"
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	ctx := context.Background()

	api := slack.New(token)
	channelID := "C0UQ8TKLJ"

	history, err := api.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     2,
	})

	if err != nil {
		panic(err)
	}

	latest := history.Messages[1].Timestamp

	fmt.Printf("latest %s\n", latest)

	history, err = api.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Oldest:    latest,
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("latest %s\n", history.Messages[0].Timestamp)

	for _, m := range history.Messages {
		fmt.Printf("> %s\n", m.Text)
	}
}
