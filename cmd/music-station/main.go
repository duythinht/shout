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
		"https://www.youtube.com/watch?v=uHio1uzzTLU",
		"https://www.youtube.com/watch?v=LjHglFd9_pc",
		"https://www.youtube.com/watch?v=gUtCfwXgDjI",
		"https://www.youtube.com/watch?v=GXQYTUdJmmI",
		"https://www.youtube.com/watch?v=vq5NvJvr55Q",
		"https://www.youtube.com/watch?v=V7_Ya16YlG8",
		"https://www.youtube.com/watch?v=NB7mpGQ46Yo",
		"https://www.youtube.com/watch?v=c5D9FbG71eE",
		"https://www.youtube.com/watch?v=sM9iSRm97Ws",
		"https://www.youtube.com/watch?v=qu7Dw4NJmY4",
		"https://www.youtube.com/watch?v=RygLJ9iToMU",
		"https://www.youtube.com/watch?v=m4xvqCmcBRU",
		"https://www.youtube.com/watch?v=hHSdja1L1XE",
		"https://www.youtube.com/watch?v=ZlL9OieDeoY",
		"https://www.youtube.com/watch?v=04pDvv3rN0g",
		"https://www.youtube.com/watch?v=3KadWjpqDXs",
		"https://www.youtube.com/watch?v=Yy4CZAj0soI",
		"https://www.youtube.com/watch?v=tz_NxOF7RB4",
		"https://www.youtube.com/watch?v=2PMnJ_Luk_o",
	}

	i := 0

	for {
		song, err := c.GetSong(ctx, playlist[i])
		if err != nil {
			tskip(err)
		}

		buff := make([]byte, 32)
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
		i = (i + 1) % len(playlist)
	}
}

func tskip(err error) {
	if err != nil {
		panic(err)
	}
}
