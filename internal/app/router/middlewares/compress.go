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

	gzipRequestReader struct {
		gzReader io.ReadCloser
	}
)

func newGZIPResponseWriter(w http.ResponseWriter) (*gzipResponseWriter, error) {
	gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &gzipResponseWriter{ResponseWriter: w, gzWriter: gw}, nil
}

func newGZIRequestReader(r *http.Request) (*gzipRequestReader, error) {
	gr, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}
	return &gzipRequestReader{gzReader: gr}, nil
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

func (gr gzipRequestReader) Read(p []byte) (n int, err error) {
	return gr.gzReader.Read(p)
}

func (gr *gzipRequestReader) Close() error {
	return gr.gzReader.Close()
}

func NewGZIPMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				reqReader, err := newGZIRequestReader(r)
				if err != nil {
					log.Errorf("failed to uncompress data, err: %s", err.Error())
					http.Error(w, "failed to uncompress data", http.StatusBadRequest)
				}
				defer reqReader.Close()
				r.Body = reqReader

			}

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
