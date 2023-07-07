package shout

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/dmulholl/mp3lib"
	"golang.org/x/exp/slog"
)

var (
	ErrWriteTimeout = errors.New("write timeout")
)

const (
	StreamDataTimeout = 200 * time.Millisecond
	ChunkFrameCount   = 50
)

type chunk struct {
	data []byte
	t    time.Duration
}

func (c *chunk) Duration() time.Duration {
	return c.t
}

type Streamer struct {
	_data  chan []byte
	_chunk chan *chunk
	r      *io.PipeReader
	w      *io.PipeWriter
}

func (s *Streamer) NextChunk() *chunk {
	return <-s._chunk
}

func OpenStreamer() *Streamer {

	r, w := io.Pipe()

	return &Streamer{
		_data:  make(chan []byte),
		_chunk: make(chan *chunk),
		r:      r,
		w:      w,
	}
}

func (s *Streamer) Write(data []byte) (int, error) {
	s._data <- data
	return len(data), nil
}

func (s *Streamer) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s *Streamer) Stream(_ context.Context) {
	go func() {
		count := 0
		for {
			var data []byte
			t := 0

			// each playback stream 0 frame
			for i := 0; i < ChunkFrameCount; i++ {
				frame := mp3lib.NextFrame(s.r)
				if frame == nil {
					continue
				}

				data = append(data, frame.RawBytes...)
				t += int(time.Second) * frame.SampleCount / frame.SamplingRate
			}

			duration := time.Duration(t) - (10 * time.Millisecond)

			slog.Info("send chunk", slog.Int("count", count))
			s._chunk <- &chunk{
				data: data,
				t:    duration,
			}

			count++
			if count < PreserveChunkCount {
				continue
			}
			slog.Info("streaming", slog.Duration("t", duration))
			time.Sleep(duration)
		}
	}()

	started := false

	for {
		select {
		case data := <-s.data():
			_, err := s.w.Write(data)
			if err != nil {
				slog.Warn("stream, write mp3", slog.String("error", err.Error()))
			}
			started = true
		case <-time.After(StreamDataTimeout):
			//slog.Warn("no stream data", slog.Duration("timeout", StreamDataTimeout))

			if !started {
				slog.Warn("not started")
				continue
			}

			s._chunk <- &chunk{
				data: nil,
				t:    StreamDataTimeout,
			}
		}
	}
}

func (s *Streamer) data() <-chan []byte {
	return s._data
}
