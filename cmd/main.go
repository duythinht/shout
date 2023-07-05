package main

import (
	"context"
	"io"
	"time"
	"webiu/radio/shout"
	"webiu/radio/utube"
)

func main() {

	ctx := context.Background()

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

	playlist := []string{
		"https://www.youtube.com/watch?v=CgNVotVutx0",
		"https://www.youtube.com/watch?v=6Q0Pd53mojY",
		"https://www.youtube.com/watch?v=RG3OEWHe1b4",
		"https://www.youtube.com/watch?v=RlTDbIutJsU",
		"https://www.youtube.com/watch?v=_8vekzCF04Q",
	}

	i := 0

	for {
		song, err := c.GetSong(ctx, playlist[i])
		if err != nil {
			panic(err)
		}

		buff := make([]byte, 32)
		for {
			n, err := song.Read(buff)
			if err != nil && err != io.EOF {
				panic(err)
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
		i = (i + 1) % len(playlist)
	}
}

func tskip(err error) {
	if err != nil {
		panic(err)
	}
}
