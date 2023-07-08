package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dmulholl/mp3lib"
)

var (
	ErrWriteTimeout = errors.New("write timeout")
)

const (
	StreamChunkedTimeout = 200 * time.Millisecond
)

type chunk struct {
	data []byte
	t    time.Duration
}

type Streamer struct {
	_chunk chan *chunk
	r      *io.PipeReader
	w      *io.PipeWriter
}

func Open() *Streamer {

	r, w := io.Pipe()

	return &Streamer{
		_chunk: make(chan *chunk),
		r:      r,
		w:      w,
	}
}

func (s *Streamer) Write(data []byte) (int, error) {
	return s.w.Write(data)
}

func (s *Streamer) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s *Streamer) Stream(_ context.Context) {
	go func() {
		for {
			var data []byte
			t := 0

			// each playback stream 0 frame
			for i := 0; i < 50; i++ {
				frame := mp3lib.NextFrame(s.r)
				if frame == nil {
					continue
				}

				data = append(data, frame.RawBytes...)
				t += int(time.Second) * frame.SampleCount / frame.SamplingRate
			}

			duration := time.Duration(t)

			s._chunk <- &chunk{
				data: data,
				t:    duration,
			}
			time.Sleep(duration)
		}
	}()
}

func (s *Streamer) NextChunk() *chunk {
	select {
	case chunked := <-s._chunk:
		return chunked
	case <-time.After(StreamChunkedTimeout):
		return &chunk{
			data: nil,
			t:    StreamChunkedTimeout,
		}
	}
}

func main() {

	files := []string{
		"1tBlaVjWwbI.mp3",
		"_8vekzCF04Q.mp3",
	}

	_ = files

	s := Open()

	go s.Stream(context.Background())

	go func() {
		for {
			chunked := s.NextChunk()
			time.Sleep(chunked.t)
			fmt.Printf("timeout %s\n", chunked.t)
		}
	}()

	//time.Sleep(10 * time.Second)

	for _, filename := range files {

		fmt.Printf("Stream %s\n", filename)

		f, err := os.Open("./songs/" + filename)

		if err != nil {
			panic(err)
		}

		io.Copy(s, f)

		time.Sleep(2 * time.Second)
	}
}
