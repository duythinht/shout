package main

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/dmulholl/mp3lib"
	"github.com/slack-go/slack"
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
	channelID := "C0UQ8TKLJ"
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	last, _ := api.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1,
	})

	api.DeleteMessage(channelID, last.Messages[0].Timestamp)
}
