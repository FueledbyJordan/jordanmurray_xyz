package renderer

import (
	"net/http"
	"strings"
)

type Renderer interface {
	Data() []byte
	CompressedData() []byte
	ContentType() string
}

func Write(w http.ResponseWriter, r *http.Request, resp Renderer) {
	w.Header().Set("Content-Type", resp.ContentType())

	if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Write(resp.CompressedData())
	} else {
		w.Write(resp.Data())
	}
}
