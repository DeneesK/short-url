package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testID = "test-id"

type ShortenerURLServiceMock struct {
	m       sync.RWMutex
	storage map[string]string
}

func (r *ShortenerURLServiceMock) ShortenURL(value string) (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	r.storage[testID] = value
	return testID, nil
}

func (r *ShortenerURLServiceMock) FindByShortened(id string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	v, ok := r.storage[id]
	if !ok {
		return "", fmt.Errorf("url not found by id: %v", id)
	}
	return v, nil
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body []byte) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	rep := &ShortenerURLServiceMock{storage: make(map[string]string)}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := *logger.Sugar()

	r := router.NewRouter(rep, &sugar)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code int
	}

	var testTable = []struct {
		name   string
		url    string
		method string
		body   []byte
		want   want
	}{
		{
			name:   "post '/'",
			url:    "/",
			method: http.MethodPost,
			body:   []byte("http://example.com"),
			want: want{
				code: http.StatusCreated,
			},
		},
		{
			name:   "get '/{id}'",
			url:    "/test-id",
			method: http.MethodGet,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "post '/' empty body",
			url:    "/",
			method: http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "get '/{id}' with wrong id",
			url:    "/wrong-id",
			method: http.MethodGet,
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, v := range testTable {
		t.Run(v.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, v.method, v.url, v.body)
			defer resp.Body.Close()
			assert.Equal(t, v.want.code, resp.StatusCode)
		})
	}
}
