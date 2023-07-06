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
	InitFrameCount = 1000
	FrameCount     = 100
)

type Shout struct {
	buffer *Buffer //buffer, for reserve data

	initialed bool
}

func New() *Shout {
	return &Shout{
		buffer: buffer(),
	}
}

func (s *Shout) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	slog.Info("Client connected", slog.String("ip", getRealIP(r)))

	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "audio/mpeg")

	init := true
	seg := 0

	for {
		bSeg := s.buffer.Segment()
		if seg == bSeg {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		if !init {
			data, t := s.buffer.Initialization()
			w.Write(data)
			time.Sleep(time.Millisecond * t)
			continue
		}
		data, t := s.buffer.Playback()
		w.Write(data)
		seg = bSeg
		time.Sleep(time.Millisecond * t)
	}
}

func (s *Shout) Stream(r io.ReadCloser) error {
	if !s.initialed {
		slog.Info("Initialize Stream for the first time")
		s.initial(r)
	}
	for {
		if err := s.playback(r); err != nil {
			return err
		}
	}
}

func (s *Shout) initial(r io.ReadCloser) error {

	var (
		idata []byte
		data  []byte
	)

	eof := false
	t := 0

	for i := 0; i < InitFrameCount; i++ {
		frame := mp3lib.NextFrame(r)
		if frame == nil {
			eof = true
			continue
		}

		idata = append(idata, frame.RawBytes...)

		if i >= InitFrameCount-FrameCount {
			data = append(data, frame.RawBytes...)
			t += 1000 * frame.SampleCount / frame.SamplingRate
		}
	}

	s.buffer.WriteInitilize(idata)

	s.buffer.Write(data, time.Duration(t))

	if eof {
		return io.EOF
	}

	s.initialed = true

	return nil
}

func (s *Shout) playback(r io.ReadCloser) error {
	var data []byte
	t := 0
	eof := false

	// each playback stream 50 frame
	for i := 0; i < FrameCount; i++ {
		frame := mp3lib.NextFrame(r)
		if frame == nil {
			eof = true
			continue
		}

		data = append(data, frame.RawBytes...)
		t += 1000 * frame.SampleCount / frame.SamplingRate
	}
	s.buffer.Write(data, time.Duration(t))
	time.Sleep(time.Duration(t) * time.Millisecond)
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
	initial  []byte
	seg      int
	t        time.Duration // current playback duration
}

func buffer() *Buffer {
	return &Buffer{
		RWMutex: &sync.RWMutex{},
	}
}

func (b *Buffer) WriteInitilize(data []byte) {
	b.Lock()
	defer b.Unlock()
	b.initial = data[:]
}

func (b *Buffer) Write(data []byte, t time.Duration) {
	b.Lock()
	defer b.Unlock()

	initial := b.initial[:]
	initial = initial[len(data):]
	initial = append(initial, data...)
	b.initial = initial
	b.playback = data[:]
	b.seg++
	b.t = t
}

func (b *Buffer) Playback() ([]byte, time.Duration) {
	b.RLock()
	defer b.RUnlock()
	return b.playback[:], b.t
}

func (b *Buffer) Initialization() ([]byte, time.Duration) {
	b.RLock()
	defer b.RUnlock()
	return b.initial[:], b.t
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
