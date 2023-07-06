package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
	"webiu/radio/shout"
	"webiu/radio/station"
	"webiu/radio/utube"
)

func main() {

	ctx := context.Background()

	token := os.Getenv("SLACK_TOKEN")
	channelID := "C0UQ8TKLJ"

	s := station.New(token, channelID)
	playlist, err := s.History(ctx)
	tskip(err)

	setNowPlaying, err := s.NowPlaying()

	tskip(err)

	cfg := &shout.Config{
		Host:     "localhost",
		Port:     8000,
		User:     "source",
		Password: "hackme",
		Mount:    "/stream.mp3",
		Proto:    shout.ProtocolHTTP,
		Format:   shout.ShoutFormatMP3,
	}

	c := utube.New("./songs/")

	w, err := shout.Connect(cfg)
	tskip(err)

	defer w.Close()
	time.Sleep(10 * time.Second)

	for {

		link := playlist.Shuffle()
		song, err := c.GetSong(ctx, link)

		if err != nil {
			tskip(err)
		}

		err = setNowPlaying(song.Video.Title)
		fmt.Printf("Now Playing: %s\n", link)

		tskip(err)

		buff := make([]byte, 1024)
		for {

			n, err := song.Read(buff)
			if err != nil && err != io.EOF {
				tskip(err)
			}
			if n == 0 {
				break
			}

			_, err = w.Write(buff)
			if err != nil {
				panic(err)
			}
		}

		//io.Copy(w, song)
		go func(io.ReadCloser) {
			time.Sleep(10 * time.Second)
			song.Close()
		}(song)

	}
}

func tskip(err error) {
	if err != nil {
		panic(err)
	}
}
