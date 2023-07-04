package main

import (
	"context"
	"fmt"
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

	m1, _ := c.GetSong(ctx, "https://www.youtube.com/watch?v=RzNbppUhS0A")
	io.Copy(w, m1)
	m1.Close()
	fmt.Println("done 1")
	time.Sleep(10 * time.Second)

	m2, _ := c.GetSong(ctx, "https://www.youtube.com/watch?v=Zq8Cy8tQr8A")
	io.Copy(w, m2)
	m2.Close()
	fmt.Println("done 2")
	time.Sleep(10 * time.Second)

	m3, _ := c.GetSong(ctx, "https://www.youtube.com/watch?v=Zq8Cy8tQr8A")
	io.Copy(w, m3)
	m3.Close()
	fmt.Println("done 3")

	time.Sleep(1 * time.Minute)
}

func tskip(err error) {
	if err != nil {
		panic(err)
	}
}
