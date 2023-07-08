package shout

import (
	"context"
	"io"
	"time"

	"github.com/dmulholl/mp3lib"
)

const (
	ReadChunkedTimeout = 200 * time.Millisecond
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

func OpenStream() *Streamer {

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

func (s *Streamer) Stream(ctx context.Context, next chan struct{}) {
	for {
		// handle next
		select {
		case <-next:
			for {

				tag := mp3lib.NextID3v2Tag(s.r)
				if tag != nil {
					break
				}
			}
		default:
		}

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

	}
}

func (s *Streamer) NextChunk() *chunk {
	select {
	case chunked := <-s._chunk:
		return chunked
	case <-time.After(ReadChunkedTimeout):
		return &chunk{
			data: nil,
			t:    ReadChunkedTimeout,
		}
	}
}
