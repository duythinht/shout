package main

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/duythinht/shout/shout"

	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()

	address := os.Getenv("SERVER_ADDRESS")

	if address == "" {
		address = "0.0.0.0:8000"
	}

	streamer := shout.OpenStream()

	shout := shout.New()
	defer shout.Close()

	go streamer.Stream(ctx)
	go shout.Attach(streamer)

	mux := chi.NewMux()

	mux.Get("/stream.mp3", shout.ServeHTTP)

	go func() {
		slog.Info("Starting server", slog.String("address", address))
		err := http.ListenAndServe(address, mux)
		qcheck(err)
	}()

	err := filepath.Walk("./songs", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		slog.Info("Streaming song...", slog.String("path", path))

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err = io.Copy(streamer, f)

		if err != nil {
			return err
		}

		slog.Info("Sleep 10 second before move to next song")
		time.Sleep(10 * time.Second)
		return nil
	})

	if err != nil {
		slog.Error("play", slog.String("error", err.Error()))
	}

}

func qcheck(err error) {
	if err != nil {
		slog.Error("station", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
