package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	gzipRequestReader struct {
		gzReader io.ReadCloser
	}
)

func newGZIRequestReader(r *http.Request) (*gzipRequestReader, error) {
	gr, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}
	return &gzipRequestReader{gzReader: gr}, nil
}

func (gr gzipRequestReader) Read(p []byte) (n int, err error) {
	return gr.gzReader.Read(p)
}

func (gr *gzipRequestReader) Close() error {
	return gr.gzReader.Close()
}

func NewRequestDecodeMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				reqReader, err := newGZIRequestReader(r)
				if err != nil {
					log.Errorf("failed to uncompress data, err: %s", err.Error())
					http.Error(w, "failed to uncompress data", http.StatusBadRequest)
				}
				defer reqReader.Close()
				r.Body = reqReader

			}

			next.ServeHTTP(w, r)
		})
	}
}
