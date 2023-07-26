package shout

import (
	"context"
	"io"
	"time"

	"github.com/dmulholl/mp3lib"
)

const (
	ReadChunkedTimeout = 50 * time.Millisecond
)

type Chunk struct {
	data    []byte
	t       time.Duration
	timeout bool
}

type stream struct {
	_chunk chan *Chunk
	r      *io.PipeReader
	w      *io.PipeWriter
}

func streamMP3() *stream {

	r, w := io.Pipe()

	return &stream{
		_chunk: make(chan *Chunk),
		r:      r,
		w:      w,
	}
}

func (s *stream) Write(data []byte) (int, error) {
	return s.w.Write(data)
}

func (s *stream) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s *stream) run(ctx context.Context, next <-chan struct{}) {
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
			data:    data,
			t:       duration,
			timeout: false,
		}

	}
}

func (s *stream) NextChunk() *Chunk {
	select {
	case chunked := <-s._chunk:
		return chunked
	case <-time.After(ReadChunkedTimeout):
		t := 0
		var data []byte

		// Make the song skip smooth by add 1 second chunk
		for t < int(time.Second) {
			t += silentDuration
			data = append(data, silentData[:]...)
		}
		return &Chunk{
			data:    data,
			t:       time.Duration(t),
			timeout: true,
		}
	}
}
