package utube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"webiu/radio/ffmpeg"

	"github.com/kkdai/youtube/v2"
	ytdl "github.com/kkdai/youtube/v2/downloader"
	"golang.org/x/net/http/httpproxy"
)

type Client struct {
	*ytdl.Downloader
}

func New(outputDir string) *Client {
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

	downloader := &ytdl.Downloader{
		OutputDir: outputDir,
	}
	downloader.Client.Debug = false
	downloader.HTTPClient = &http.Client{Transport: httpTransport}

	return &Client{downloader}
}

func (c *Client) GetSong(ctx context.Context, link string) (io.ReadCloser, error) {

	video, err := c.GetVideo(link)

	fmt.Println("Now Playing: ", video.Title)

	if err != nil {
		return nil, fmt.Errorf("get video info %w", err)
	}

	webm := fmt.Sprintf("%s/%s.webm", c.OutputDir, video.ID)
	mp3 := fmt.Sprintf("%s/%s.mp3", c.OutputDir, video.ID)

	_, err = os.Stat(c.OutputDir + mp3)

	if errors.Is(err, os.ErrNotExist) {

		fm, err := getAudioWebmFormat(video)

		if err != nil {
			return nil, fmt.Errorf("get video info %w", err)
		}

		err = c.Download(ctx, video, fm, video.ID+".webm")
		if err != nil {
			return nil, fmt.Errorf("download %w", err)
		}

		ffmpeg.ToMP3(
			ctx,
			webm,
			mp3,
		)

		os.Remove(webm)
	}
	return os.Open(mp3)
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
