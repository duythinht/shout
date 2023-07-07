package shout

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dmulholl/mp3lib"
	"golang.org/x/exp/slog"
)

const (
	ChunkFrameCount    = 50
	PreserveChunkCount = 5
)

type Shout struct {
	//*Buffer //buffer, for reserve data
	Buffer *atomic.Pointer[Buffer]
	init   bool
}

func New() *Shout {

	buf := &atomic.Pointer[Buffer]{}
	buf.Store(&Buffer{})

	return &Shout{
		Buffer: buf,
		init:   false,
	}
}

func (s *Shout) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ip := getRealIP(r)

	slog.Info("Client connected", slog.String("ip", ip), slog.String("path", r.URL.Path))

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Keep-Alive
	w.Header().Set("Connection", "Keep-Alive")

	// Cache-Control
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "audio/mpeg")

	seg := 0

	ctx := r.Context()

	init := false

	for {
		select {
		case <-ctx.Done():
			slog.Info("Client disconnected", slog.String("ip", ip), slog.String("path", r.URL.Path))
			return
		default:
		}

		b := s.Buffer.Load()

		if seg == b.seg {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		if !init {
			w.Write(b.playback)
			init = true
		} else {
			_, err := w.Write(b.playback[len(b.playback)-b.lenght:])
			if err != nil {
				slog.Warn("Client disconnected", slog.String("ip", ip), slog.String("path", r.URL.Path))
				return
			}
		}

		seg = b.seg
		time.Sleep(b.t)
	}
}

func (s *Shout) StreamAll(r io.ReadCloser) error {

	if !s.init {
		err := s.initialize(r)
		if err != nil {
			return err
		}
		s.init = true
	}

	for {

		chunked, t, err := nextchunk(r)

		buf := s.Buffer.Load()

		playback := buf.playback[len(chunked):]

		s.Buffer.Store(&Buffer{
			playback: append(playback, chunked...),
			seg:      buf.seg + 1,
			t:        t,
			lenght:   len(chunked),
		})

		time.Sleep(t)

		// usual return when EOF
		if err != nil {
			return err
		}
	}
}

func (s *Shout) initialize(r io.ReadCloser) error {

	slog.Info("Initilize a stream", slog.Int("preserve-chunk-count", PreserveChunkCount))

	var (
		playback []byte
		chunked  []byte
		t        time.Duration
		err      error
		lenght   int
	)

	for i := 0; i < PreserveChunkCount; i++ {
		chunked, t, err = nextchunk(r)

		if err != nil {
			return err
		}

		playback = append(playback, chunked...)
		lenght = len(chunked)
	}

	s.Buffer.Store(&Buffer{
		playback: playback[:],
		seg:      1,
		t:        t,
		lenght:   lenght,
	})

	time.Sleep(t)

	return nil
}

func nextchunk(r io.ReadCloser) ([]byte, time.Duration, error) {
	var data []byte
	t := 0

	// each playback stream 0 frame
	for i := 0; i < ChunkFrameCount; i++ {
		frame := mp3lib.NextFrame(r)
		if frame == nil {
			return data, time.Duration(t), io.EOF
		}

		data = append(data, frame.RawBytes...)
		t += int(time.Second) * frame.SampleCount / frame.SamplingRate
	}

	return data, time.Duration(t), nil
}

func (s *Shout) Close() error {
	return nil
}

type Buffer struct {
	playback []byte
	seg      int
	t        time.Duration // current playback duration
	lenght   int
}

func getRealIP(r *http.Request) string {
	xfwd4 := r.Header.Get("X-Forwarded-For")

	if xfwd4 == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}

	ips := strings.Split(xfwd4, ", ")
	return ips[len(ips)-1]
}
