package shout

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dmulholl/mp3lib"
	"golang.org/x/exp/slog"
)

const (
	ChunkFrameCount = 100
)

type Shout struct {
	*Buffer //buffer, for reserve data
}

func New() *Shout {
	return &Shout{
		Buffer: buffer(),
	}
}

func (s *Shout) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ip := getRealIP(r)

	slog.Info("Client connected", slog.String("ip", ip), slog.String("path", r.URL.Path))

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Keep-Alive??
	// w.Header().Set("Connection", "Keep-Alive")

	// Cache-Control
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "audio/mpeg")

	seg := 0

	ctx := r.Context()

	for {

		select {
		case <-ctx.Done():
			slog.Info("Client disconnected", slog.String("ip", ip), slog.String("path", r.URL.Path))
			return
		default:
		}

		bSeg := s.Segment()
		if seg == bSeg {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		data, duration := s.Playback()
		_, err := w.Write(data)
		if err != nil {
			slog.Warn("Client disconnected", slog.String("ip", ip), slog.String("path", r.URL.Path))
			return
		}
		seg = bSeg
		time.Sleep(duration)
	}
}

func (s *Shout) StreamAll(r io.ReadCloser) error {

	for {
		chunked, t, err := s.nextChunk(r)

		// alway write chunked
		s.Write(chunked, t)

		time.Sleep(t)

		// usual return when EOF
		if err != nil {
			return err
		}
	}
}

func (s *Shout) nextChunk(r io.ReadCloser) ([]byte, time.Duration, error) {
	var data []byte
	t := 0

	// each playback stream 50 frame
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
	*sync.RWMutex

	playback []byte
	seg      int
	t        time.Duration // current playback duration
}

func buffer() *Buffer {
	return &Buffer{
		RWMutex: &sync.RWMutex{},
	}
}

func (b *Buffer) Write(data []byte, t time.Duration) {
	b.Lock()
	defer b.Unlock()

	b.playback = data[:]
	b.seg++
	b.t = t
}

func (b *Buffer) Playback() ([]byte, time.Duration) {
	b.RLock()
	defer b.RUnlock()
	return b.playback[:], b.t
}

func (b *Buffer) Segment() int {
	b.RLock()
	defer b.RUnlock()
	return b.seg
}

func getRealIP(r *http.Request) string {
	xfwd4 := r.Header.Get("X-Forwarded-For")

	if xfwd4 == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}

	ips := strings.Split(xfwd4, ", ")
	return ips[len(ips)-1]
}
