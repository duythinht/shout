package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/duythinht/shout/utube"

	"github.com/duythinht/shout/station"

	"github.com/duythinht/shout/shout"

	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("SLACK_TOKEN")
	channelID := "C0UQ8TKLJ"

	s := station.New(token, channelID)
	playlist, err := s.History(ctx)
	qcheck(err)

	setTitle, err := s.NowPlaying()

	qcheck(err)

	c := utube.New("./songs/")

	w := shout.New()

	defer w.Close()

	go func() {
		slog.Info("Starting server at 0.0.0.0:8000")
		http.ListenAndServe("0.0.0.0:8000", w)
	}()

	for {

		link := playlist.Shuffle()
		song, err := c.GetSong(ctx, link)

		if err != nil {
			if errors.Is(err, utube.ErrSongTooLong) {
				continue
			}
			qcheck(err)
		}

		title := song.Video.Title

		slog.Info("Now Playing", slog.String("link", link), slog.String("title", title))

		err = setTitle(song.Video.Title)

		qcheck(err)

		err = w.StreamAll(song)

		if err != nil && !errors.Is(err, io.EOF) {
			qcheck(err)
		}

		song.Close()
	}
}

func qcheck(err error) {
	if err != nil {
		slog.Error("station", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
