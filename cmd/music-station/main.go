package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/duythinht/shout/utube"
	"github.com/duythinht/shout/web"
	"github.com/go-chi/chi/v5"

	"github.com/duythinht/shout/station"

	"github.com/duythinht/shout/shout"

	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

//

func main() {
	token := os.Getenv("SLACK_TOKEN")
	address := os.Getenv("SERVER_ADDRESS")

	ctx := context.Background()
	next := make(chan struct{})

	if address == "" {
		address = "0.0.0.0:8000"
	}

	channelID := "C0UQ8TKLJ"

	st := station.New(token, channelID)

	playlist, err := st.History(ctx)
	qcheck(err)

	queue, err := st.Watch(ctx)
	qcheck(err)

	youtube := utube.New("./songs/")

	sh := shout.New()
	defer sh.Close()

	go sh.Streaming(ctx, next)

	mux := chi.NewMux()

	mux.Get("/", web.ServeIndex)
	mux.Get("/stream.mp3", sh.ServeHTTP)

	var title atomic.Value

	mux.Get("/now-playing", websocket.Handler(func(ws *websocket.Conn) {

		var clientTitle string

		one := time.Tick(1 * time.Second)
		fifteen := time.Tick(15 * time.Second)

		for {
			select {
			case <-ws.Request().Context().Done():
				return
			case <-one:
				currentTitle := title.Load().(string)
				if clientTitle != currentTitle {
					clientTitle = currentTitle
					payload, _ := json.Marshal(map[string]string{
						"title": currentTitle,
					})

					ws.Write(payload)
				}
			case <-fifteen:
				_, _ = ws.Write([]byte{'\n'})
			}
		}

	}).ServeHTTP)

	mux.Get("/list.txt", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		fmt.Fprintf(w, "# Songs in Queuing\n\n")
		for _, link := range queue.Links() {
			title, err := youtube.GetSongTitle(link)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "%s - %s\n", link, title)
		}
	})

	mux.Post("/next", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Song skip requested")
		next <- struct{}{}
	})

	go func() {
		slog.Info("Starting server", slog.String("address", address))
		err := http.ListenAndServe(address, mux)
		qcheck(err)
	}()

	onAir, stop := st.OnAir()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		stop()
		os.Exit(0)
	}()

	for {

		var link string

		if queue.Size() > 0 {
			// Get the link from queue, and then add it back to playlist for play later (when queue is empty)
			link = queue.Poll()
			playlist.Add(link)
		} else {
			// play shuffle if shout don't have any music in queue
			link = playlist.Shuffle()
		}

		song, err := youtube.GetSong(ctx, link)

		if err != nil {
			// skip this song if ytb get some error
			slog.Warn("get song", slog.String("error", err.Error()))
			continue
		}

		slog.Info("Now Playing", slog.String("link", link), slog.String("title", song.Video.Title))

		err = st.SetBookmark(song.Video.Title)
		qcheck(err)

		onAir(ctx, song.Video.Title, link)

		title.Store(song.Video.Title)

		_, err = io.Copy(sh, song)

		if err != nil && !errors.Is(err, io.EOF) {
			qcheck(err)
		}

		if err := song.Close(); err != nil {
			return
		}
	}
}

// qcheck to support quick check and exit
func qcheck(err error) {
	if err != nil {
		slog.Error("station", err)
		os.Exit(1)
	}
}
