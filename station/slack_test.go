package station_test

import (
	"errors"
	"testing"

	"github.com/duythinht/shout/station"
)

func TestExtractVideoFromMessage(t *testing.T) {
	id1, err := station.ExtractYoutubeID("https://m.youtube.com/watch?v=c5D9FbG71eE&amp;t=144s|https://m.youtube.com/watch?v=c5D9FbG71eE")

	if err != nil {
		t.Fatal(err)
	}

	if id1 != "c5D9FbG71eE" {
		t.Fail()
	}

	id2, _ := station.ExtractYoutubeID("https://www.youtube.com/watch?v=Yy4CZAj0soI")
	if id2 != "Yy4CZAj0soI" {
		t.Fail()
	}

	id3, err := station.ExtractYoutubeID("https://www.youtube.com/abc")

	if !errors.Is(err, station.ErrNotYoutubeLink) {
		t.Logf("Err %s - %s", id3, err.Error())
		t.Fail()
	}
}
