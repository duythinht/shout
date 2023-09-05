package web

import (
	_ "embed"
	"net/http"
)

//go:embed index.html
var index []byte

func ServeIndex(w http.ResponseWriter, _ *http.Request) {
	w.Write(index)
}
