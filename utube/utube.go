package utube

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/duythinht/shout/ffmpeg"

	"github.com/kkdai/youtube/v2"
	"golang.org/x/net/http/httpproxy"
)

var (
	ErrSongTooLong = errors.New("video too long, assume this a troll")
)

type Client struct {
	*youtube.Client
	dir        string
	songTitles *sync.Map
}

type Song struct {
	io.Reader
	Video *youtube.Video
}

func (s *Song) Close() error {
	return nil
}

func New(songDirectory string) *Client {
	proxyFunc := httpproxy.FromEnvironment().ProxyFunc()
	httpTransport := &http.Transport{
		// Proxy: http.ProxyFromEnvironment() does not work. Why?
		Proxy: func(r *http.Request) (uri *url.URL, err error) {
			return proxyFunc(r.URL)
		},
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Client{
		Client: &youtube.Client{
			HTTPClient: &http.Client{Transport: httpTransport},
		},
		dir:        songDirectory,
		songTitles: &sync.Map{},
	}
}

func (c *Client) GetSongTitle(link string) (string, error) {
	if title, ok := c.songTitles.Load(link); ok {
		return title.(string), nil
	}

	video, err := c.GetVideo(link)

	if err != nil {
		return "", fmt.Errorf("get video info %w", err)
	}

	c.songTitles.Store(link, video.Title)

	return video.Title, nil
}

func (c *Client) GetSong(ctx context.Context, link string) (*Song, error) {

	video, err := c.GetVideo(link)

	if err != nil {
		return nil, fmt.Errorf("get video info %w", err)
	}

	if video.Duration > 13*time.Minute {
		return nil, fmt.Errorf("video duration too long %s %w", link, ErrSongTooLong)
	}

	fm, err := getAudioWebmFormat(video)

	if err != nil {
		return nil, fmt.Errorf("get video info %w", err)
	}

	stream, total, err := c.GetStream(video, fm)

	if err != nil {
		return nil, fmt.Errorf("download - get stream - %w", err)
	}

	defer stream.Close()

	data, err := io.ReadAll(stream)

	if err != nil {
		return nil, fmt.Errorf("download - stream - %w", err)
	}

	mp3, err := ffmpeg.WebmToMp3(ctx, bytes.NewReader(data))

	if err != nil {
		return nil, fmt.Errorf("convert webm to mp3 %w, size: %d, len: %d", err, total, len(data))
	}

	return &Song{
		Video:  video,
		Reader: mp3,
	}, nil
}

func getAudioWebmFormat(v *youtube.Video) (*youtube.Format, error) {
	formats := v.Formats

	audioFormats := formats.Type("audio")
	audioFormats.Sort()
	for _, fm := range formats {
		if strings.HasPrefix(fm.MimeType, "audio/webm") {
			//slog.Info("get webm format", "title", v.Title, "url", fm.URL)
			return &fm, nil
		}
	}
	// no webm, take first format
	return &formats[0], nil
}
