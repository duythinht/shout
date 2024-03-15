package web

import (
	_ "embed"
	"net/http"
)

//go:embed index.html
var index []byte

//go:embed css/main.css
var css []byte

//go:embed scripts/script.js
var script []byte

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	switch r.URL.Path {
	case "/css/main.css":
		w.Header().Set("Content-Type", "text/css")
		w.Write(css)
	case "/scripts/script.js":
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(script)
	default:
		w.Header().Set("Content-Type", "text/html")
		w.Write(index)
	}
}
