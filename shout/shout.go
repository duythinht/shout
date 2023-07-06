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

	slog.Info("Client connected", slog.String("ip", getRealIP(r)))

	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "audio/mpeg")

	seg := 0

	for {
		bSeg := s.Segment()
		if seg == bSeg {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		data, duration := s.Playback()
		w.Write(data)
		seg = bSeg
		time.Sleep(duration)
	}
}

func (s *Shout) StreamAll(r io.ReadCloser) error {
	for {
		if err := s.stream(r); err != nil {
			return err
		}
	}
}

func (s *Shout) stream(r io.ReadCloser) error {
	var data []byte
	t := 0
	eof := false

	// each playback stream 50 frame
	for i := 0; i < ChunkFrameCount; i++ {
		frame := mp3lib.NextFrame(r)
		if frame == nil {
			eof = true
			continue
		}

		data = append(data, frame.RawBytes...)
		t += 1000 * frame.SampleCount / frame.SamplingRate
	}

	// duration of buffer will be reduce 5ms for ensure stream gap
	duration := time.Duration(t) * time.Millisecond
	s.Write(data, duration)
	time.Sleep(duration)
	if eof {
		return io.EOF
	}
	return nil
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
