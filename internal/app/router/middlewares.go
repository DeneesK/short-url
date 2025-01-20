package router

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	compressibleContentTypes = "application/json;text/html"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	gzipResponseWriter struct {
		http.ResponseWriter
		GzWriter *gzip.Writer
	}

	gzipRequestReader struct {
		GzReader io.ReadCloser
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	switch {
	case strings.Contains(compressibleContentTypes, contentType):
		return g.GzWriter.Write(b)
	default:
		g.GzWriter.Reset(io.Discard)
		return g.ResponseWriter.Write(b)
	}
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := g.Header().Get("Content-Type")
	if !strings.Contains(compressibleContentTypes, contentType) {
		g.Header().Del("Content-Encoding")
	}
	g.ResponseWriter.WriteHeader(statusCode)
}

func (gr gzipRequestReader) Read(p []byte) (n int, err error) {
	return gr.GzReader.Read(p)
}

func (gr *gzipRequestReader) Close() error {
	return gr.GzReader.Close()
}

func NewLoggingMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			log.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)
		})
	}
}

func NewGZIPMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				gr, err := gzip.NewReader(r.Body)
				if err != nil {
					log.Errorf("failed to uncompress data, err: %s", err.Error())
					http.Error(w, "failed to uncompress data", http.StatusBadRequest)
				}
				defer gr.Close()

				r.Body = &gzipRequestReader{GzReader: gr}
			}

			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					log.Errorf("failed to compress data, err: %s", err.Error())
					http.Error(w, "failed to compress data", http.StatusBadRequest)
				}
				defer gw.Close()

				w.Header().Set("Content-Encoding", "gzip")
				ow = &gzipResponseWriter{ResponseWriter: w, GzWriter: gw}
			}

			next.ServeHTTP(ow, r)
		})
	}
}
