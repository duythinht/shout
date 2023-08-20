package ffmpeg_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/duythinht/shout/ffmpeg"
)

func TestConvertWebmToMp3(t *testing.T) {
	f, err := os.Open("../songs/test.webm")

	if err != nil {
		t.Logf("error %s", err)
		t.Fail()
	}

	out, err := ffmpeg.WebmToMp3(context.Background(), f)

	if err != nil {
		t.Logf("error %s", err)
		t.Fail()
	}

	data, err := io.ReadAll(out)

	if err != nil {
		t.Logf("error %s", err)
		t.Fail()
	}

	if len(data) < 100 {
		t.Logf("len %d", len(data))
		t.Fail()
	}
}
