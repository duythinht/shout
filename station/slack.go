package station

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/slack-go/slack"
	"golang.org/x/exp/slog"
)

var (
	rx = regexp.MustCompile(`https://(.+.youtube.com|youtu.be)/(watch\?v=([^&^>^|]+)|([^&^>^|]+))`)

	//hardcode stream link, move to config later
	streamLink = "https://radio.0x97a.com/stream.mp3"

	ErrNotYoutubeLink = errors.New("not a youtube link")
)

// ExtractYoutubeID from message
func ExtractYoutubeID(message string) (id string, err error) {
	if rx.MatchString(message) {

		sub := rx.FindStringSubmatch(message)
		if len(sub) > 3 {

			// whenever link is youtu.be return sub 4
			if sub[1] == "youtu.be" && len(sub[4]) > 0 {
				return sub[4], nil
			}

			// otherwise, assume link is youtube.com, take the id from ?v=... and check the len

			if len(sub[3]) > 0 {
				return sub[3], nil
			}
		}
	}

	// otherwise assume that not a valid youtube link
	return "", fmt.Errorf("%s %w", message, ErrNotYoutubeLink)
}

// Station of #music channel
type Station struct {
	*slack.Client
	channelID string
}

// New return station by slack token and channel
func New(slackToken string, channelID string) (station *Station) {
	api := slack.New(slackToken)

	return &Station{
		Client:    api,
		channelID: channelID,
	}
}

func (s Station) History(ctx context.Context) (playlist *Playlist, err error) {
	playlist = NewPlaylist()

	var (
		cursor = ""
		more   = true
	)

	for more {
		history, err := s.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
			ChannelID: s.channelID,
			Cursor:    cursor,
		})

		if err != nil {
			return nil, fmt.Errorf("station history %w", err)
		}

		more = history.HasMore
		cursor = history.ResponseMetaData.NextCursor

		for _, msg := range history.Messages {
			id, err := ExtractYoutubeID(msg.Text)
			if err != nil {
				if errors.Is(err, ErrNotYoutubeLink) {
					continue
				}
				return nil, fmt.Errorf("station history, extract id %w", err)
			}
			playlist.Add(
				fmt.Sprintf(
					"https://www.youtube.com/watch?v=%s",
					id,
				),
			)
		}

	}
	return playlist, nil
}

func (s *Station) NowPlaying() (func(string) error, error) {
	bookmarks, err := s.ListBookmarks(s.channelID)

	if err != nil {
		panic(err)
	}

	for i := range bookmarks {
		bookmark := bookmarks[i]
		if bookmark.Link == streamLink {
			return func(title string) error {
				playingTitle := fmt.Sprintf("Now Playing: %s", title)
				_, err := s.EditBookmark(s.channelID, bookmark.ID, slack.EditBookmarkParameters{
					Title: &playingTitle,
				})
				return err
			}, nil
		}
	}

	bookmark, err := s.AddBookmark(s.channelID, slack.AddBookmarkParameters{
		Title: "Now Playing: -",
		Link:  streamLink,
		Emoji: ":studio_microphone:",
	})

	if err != nil {
		return nil, err
	}

	return func(title string) error {
		playingTitle := fmt.Sprintf("Now Playing: %s", title)
		_, err := s.EditBookmark(s.channelID, bookmark.ID, slack.EditBookmarkParameters{
			Title: &playingTitle,
		})
		return err
	}, nil
}

func (s *Station) Watch(ctx context.Context) (playlist *Playlist, err error) {
	playlist = NewPlaylist()

	last, err := s.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: s.channelID,
		Limit:     1,
	})

	if err != nil {
		return nil, fmt.Errorf("watch - last history %w", err)
	}

	go func() {
		ts := last.Messages[0].Timestamp
		for {

			time.Sleep(30 * time.Second)

			last, err := s.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
				ChannelID: s.channelID,
				Oldest:    ts,
			})

			if err != nil {
				slog.Warn("watch", slog.String("error", err.Error()))
				continue
			}

			count := len(last.Messages)

			if count < 1 {
				continue
			}

			for i := 0; i < count; i++ {
				text := last.Messages[count-i-1].Text
				id, err := ExtractYoutubeID(text)
				if err != nil {
					if errors.Is(err, ErrNotYoutubeLink) {
						continue
					}
					slog.Warn("watch", slog.String("error", err.Error()))
					continue
				}

				link := fmt.Sprintf(
					"https://www.youtube.com/watch?v=%s",
					id,
				)

				slog.Info("Queue added", slog.String("link", link))

				playlist.Add(link)
			}

			ts = last.Messages[0].Timestamp
		}
	}()

	return playlist, nil
}
