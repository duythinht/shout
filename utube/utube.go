package utube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/duythinht/shout/ffmpeg"

	"github.com/kkdai/youtube/v2"
	"golang.org/x/exp/slog"
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
	*os.File
	Video *youtube.Video
}

func (s *Song) Close() error {
	defer os.Remove(s.File.Name())
	return s.File.Close()
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

	mp3 := filepath.Join(c.dir, fmt.Sprintf("%s.mp3", video.ID))

	_, err = os.Stat(mp3)

	if errors.Is(err, os.ErrNotExist) {

		fm, err := getAudioWebmFormat(video)

		if err != nil {
			return nil, fmt.Errorf("get video info %w", err)
		}

		origin := filepath.Join(c.dir, fmt.Sprintf("%s.%s", video.ID, pickIdealFileExtension(fm.MimeType)))

		stream, _, err := c.GetStream(video, fm)

		if err != nil {
			return nil, fmt.Errorf("download - get stream - %w", err)
		}

		f, err := os.Create(origin)

		if err != nil {
			return nil, fmt.Errorf("download - create origin - %w", err)
		}

		_, err = io.Copy(f, stream)

		if err != nil {
			return nil, fmt.Errorf("download - io.Copy - %w", err)
		}

		slog.Info("Fetch Song From youtube", slog.String("origin", origin), slog.String("mp3", mp3))

		err = ffmpeg.ToMP3(
			ctx,
			origin,
			mp3,
		)

		if err != nil {
			return nil, fmt.Errorf("ffmpeg - %w", err)
		}

		err = os.Remove(origin)

		if err != nil {
			return nil, fmt.Errorf("remove origin - %w", err)
		}
	}

	f, err := os.Open(mp3)
	if err != nil {
		return nil, fmt.Errorf("open mp3 file %w", err)
	}

	return &Song{
		Video: video,
		File:  f,
	}, nil
}

func getAudioWebmFormat(v *youtube.Video) (*youtube.Format, error) {
	formats := v.Formats

	audioFormats := formats.Type("audio")
	audioFormats.Sort()
	for _, fm := range formats {
		if strings.HasPrefix(fm.MimeType, "audio/webm") {
			return &fm, nil
		}
	}
	// no webm, take first format
	return &formats[0], nil
}

var canonicals = map[string]string{
	"video/quicktime":  ".mov",
	"video/x-msvideo":  ".avi",
	"video/x-matroska": ".mkv",
	"video/mpeg":       ".mpeg",
	"video/webm":       ".webm",
	"video/3gpp2":      ".3g2",
	"video/x-flv":      ".flv",
	"video/3gpp":       ".3gp",
	"video/mp4":        ".mp4",
	"video/ogg":        ".ogv",
	"video/mp2t":       ".ts",
}

func pickIdealFileExtension(mediaType string) string {
	mediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "webm"
	}

	if extension, ok := canonicals[mediaType]; ok {
		return extension
	}

	// Our last resort is to ask the operating system, but these give multiple results and are rarely canonical.
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || extensions == nil {
		return "webm"
	}

	return extensions[0]
}
