package station

import (
	"errors"
	"testing"
)

func TestExtractVideoFromMessage(t *testing.T) {
	id1, err := ExtractYoutubeID("https://m.youtube.com/watch?v=c5D9FbG71eE&amp;t=144s|https://m.youtube.com/watch?v=c5D9FbG71eE")

	if err != nil {
		t.Fatal(err)
	}

	if id1 != "c5D9FbG71eE" {
		t.Fail()
	}

	id2, err := ExtractYoutubeID("https://www.youtube.com/watch?v=Yy4CZAj0soI")
	if id2 != "Yy4CZAj0soI" {
		t.Fail()
	}

	id3, err := ExtractYoutubeID("https://www.youtube.com/abc")

	if !errors.Is(err, ErrNotYoutubeLink) {
		t.Logf("Err %s - %s", id3, err)
		t.Fail()
	}
}
