package station_test

import (
	"testing"

	"github.com/duythinht/shout/station"
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

	p.Add("a")
	p.Add("b")
	p.Add("c")

	link := p.Poll()

	if p.Has("a") {
		t.Errorf("item xyz should removed")
	}

	if link != "a" {
		t.Errorf("item should be xyz, got %s", link)
	}

	if p.Size() != 2 {
		t.Errorf("size should be %d", p.Size())
	}

	links := p.Links()
	for i, link := range []string{"b", "c"} {
		if links[i] != link {
			t.Errorf("not match link %s %s", links[i], link)
		}
	}
}
