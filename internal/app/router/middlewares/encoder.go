package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	compressibleContentTypes = "application/json;text/html"
)

type (
	gzipResponseWriter struct {
		http.ResponseWriter
		gzWriter *gzip.Writer
	}
)

func newGZIPResponseWriter(w http.ResponseWriter) (*gzipResponseWriter, error) {
	gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &gzipResponseWriter{ResponseWriter: w, gzWriter: gw}, nil
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	switch {
	case strings.Contains(compressibleContentTypes, contentType):
		return g.gzWriter.Write(b)
	default:
		g.gzWriter.Reset(io.Discard)
		return g.ResponseWriter.Write(b)
	}
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := g.Header().Get("Content-Type")
	if strings.Contains(compressibleContentTypes, contentType) {
		g.Header().Set("Content-Encoding", "gzip")
	}
	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Close() error {
	return g.gzWriter.Close()
}

func NewResponseEncodeMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				rw, err := newGZIPResponseWriter(w)
				if err != nil {
					log.Errorf("failed to compress data, err: %s", err.Error())
					http.Error(w, "failed to compress data", http.StatusBadRequest)
				}
				defer rw.Close()
				ow = rw
			}

			next.ServeHTTP(ow, r)
		})
	}
}
