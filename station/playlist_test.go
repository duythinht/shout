package station_test

import (
	"testing"
	"webiu/radio/station"
)

func TestPlaylistAddAndRemove(t *testing.T) {
	p := station.NewPlaylist()

	p.Add("abc")
	p.Add("xyz")
	p.Add("xyz")

	if p.Size() != 2 {
		t.Errorf("size should be 2 after insert 2 item")
	}

	p.Delete("abc")

	if p.Size() != 1 {
		t.Errorf("size should be 1 after delete 1 item, got %d", p.Size())
	}

	if p.Has("abc") {
		t.Errorf("item abc should removed")
	}
}

func TestPlaylistPoll(t *testing.T) {
	p := station.NewPlaylist()

	p.Add("abc")
	p.Add("xyz")

	link := p.Poll()

	if p.Has("xyz") {
		t.Errorf("item xyz should removed")
	}

	if link != "xyz" {
		t.Errorf("item should be xyz, got %s", link)
	}

	if p.Size() != 1 {
		t.Errorf("size should be %d", p.Size())
	}
}
