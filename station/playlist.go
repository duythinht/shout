package station

import (
	"sync"
	"time"

	"golang.org/x/exp/rand"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

type Playlist struct {
	lock  *sync.RWMutex
	links []string
	m     map[string]struct{}
}

func NewPlaylist() *Playlist {
	return &Playlist{
		lock:  &sync.RWMutex{},
		links: make([]string, 0),
		m:     make(map[string]struct{}, 0),
	}
}

func (p *Playlist) Add(link string) {

	if _, ok := p.m[link]; ok {
		return
	}

	p.links = append(p.links, link)
	p.m[link] = struct{}{}
}

func (p *Playlist) Delete(link string) {
	if _, ok := p.m[link]; !ok {
		return
	}

	maps.DeleteFunc(p.m, func(k string, v struct{}) bool {
		return k == link
	})
	i := slices.Index(p.links, link)
	p.links = slices.Delete(p.links, i, i+1)
}

func (p *Playlist) Has(link string) bool {
	_, ok := p.m[link]
	return ok
}

func (p *Playlist) Size() int {
	return len(p.links)
}

// Poll return the fist link, if len == 0 return ""
func (p *Playlist) Poll() (link string) {
	lenght := len(p.links)

	if lenght == 0 {
		return ""
	}

	link = p.links[lenght-1]

	p.links = slices.Delete(p.links, lenght-1, lenght)
	maps.DeleteFunc(p.m, func(k string, v struct{}) bool {
		return k == link
	})
	return
}

func (p *Playlist) Shuffle() (link string) {
	i := rand.Intn(len(p.links))
	return p.links[i]
}
