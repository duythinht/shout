package shout

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slog"
)

const (
	PreserveChunkCount = 3
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
				slog.Warn(
					"Client disconnected",
					slog.String("ip", ip),
					slog.String("path", r.URL.Path),
					slog.String("error", err.Error()),
				)
				return
			}
		}

		seg = b.seg
		time.Sleep(b.t)
	}
}

// Attach to a streamer
func (s *Shout) Attach(st *Streamer) {

	var (
		playback []byte
		chunked  *Chunk
	)

	// Reserving buffer

	count := PreserveChunkCount

	for count > 0 {
		chunked = st.NextChunk()
		if len(chunked.data) > 0 {
			playback = append(playback, chunked.data...)
			count--
		}
		time.Sleep(50 * time.Millisecond)
	}

	slog.Info(
		"Init playback",
		slog.Int("playback-len", len(playback)),
		slog.Int("chunk-len", len(chunked.data)),
	)

	// Send init playback buffer
	s.Buffer.Store(&Buffer{
		playback: playback,
		seg:      0,
		t:        chunked.t,
		lenght:   len(chunked.data),
	})

	// Then start stream

	for {
		chunked := st.NextChunk()

		buf := s.Buffer.Load()

		playback := buf.playback[len(chunked.data):]

		s.Buffer.Store(&Buffer{
			playback: append(playback, chunked.data...),
			seg:      buf.seg + 1,
			t:        chunked.t,
			lenght:   len(chunked.data),
		})
		time.Sleep(chunked.t)
	}
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
