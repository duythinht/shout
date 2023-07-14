package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/duythinht/shout/utube"
	"github.com/go-chi/chi/v5"

	"github.com/duythinht/shout/station"

	"github.com/duythinht/shout/shout"

	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	address := os.Getenv("SERVER_ADDRESS")

	ctx := context.Background()
	next := make(chan struct{})

	if address == "" {
		address = "0.0.0.0:8000"
	}

	channelID := "C0UQ8TKLJ"

	station := station.New(token, channelID)

	playlist, err := station.History(ctx)
	qcheck(err)

	youtube := utube.New("./songs/")
	streamer := shout.OpenStream()

	shout := shout.New()
	defer shout.Close()

	go streamer.Stream(ctx, next)
	go shout.Attach(streamer)

	mux := chi.NewMux()

	mux.Get("/stream.mp3", shout.ServeHTTP)

	var title atomic.Value

	mux.Get("/now-playing", websocket.Handler(func(ws *websocket.Conn) {

		var clientTitle string

		for {
			select {
			case <-ws.Request().Context().Done():
				return
			case <-time.After(1 * time.Second):
				currentTitle := title.Load().(string)
				if clientTitle != currentTitle {
					clientTitle = currentTitle
					payload, _ := json.Marshal(map[string]string{
						"title": currentTitle,
					})
					ws.Write(payload)
				}
			}
		}

	}).ServeHTTP)

	mux.Post("/next", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Song skip requested")
		next <- struct{}{}
	})

	go func() {
		slog.Info("Starting server", slog.String("address", address))
		err := http.ListenAndServe(address, mux)
		qcheck(err)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		os.Exit(0)
	}()

	for {

		link := playlist.Shuffle()

		song, err := youtube.GetSong(ctx, link)

		if err != nil {
			// skip this song if client get some error
			slog.Warn("get song", slog.String("error", err.Error()))
			continue
		}

		slog.Info("Now Playing", slog.String("link", link), slog.String("title", song.Video.Title))
		title.Store(song.Video.Title)

		_, err = io.Copy(streamer, song)

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
