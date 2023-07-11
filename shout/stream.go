package shout

import (
	"context"
	"io"
	"time"

	"github.com/dmulholl/mp3lib"
)

const (
	ReadChunkedTimeout = 100 * time.Millisecond
)

type Chunk struct {
	data []byte
	t    time.Duration
}

type Streamer struct {
	_chunk chan *Chunk
	r      *io.PipeReader
	w      *io.PipeWriter
}

func OpenStream() *Streamer {

	r, w := io.Pipe()

	return &Streamer{
		_chunk: make(chan *Chunk),
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

		s._chunk <- &Chunk{
			data: data,
			t:    duration,
		}

	}
}

func (s *Streamer) NextChunk() *Chunk {
	select {
	case chunked := <-s._chunk:
		return chunked
	case <-time.After(ReadChunkedTimeout):
		t := 0
		var data []byte
		for t < int(ReadChunkedTimeout) {
			t += silentDuration
			data = append(data, silentData[:]...)
		}
		return &Chunk{
			data: data,
			t:    time.Duration(t),
		}
	}
}
