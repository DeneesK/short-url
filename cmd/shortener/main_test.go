package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DeneesK/short-url/internal/app/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rep struct {
}

func (r *rep) SaveURL(url string) (string, error) {
	return "id", nil
}

func (r *rep) GetURL(id string) (string, error) {
	return "example.com", nil
}

func TestURLHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		urlPath string
		body    []byte
		method  string
		want    want
		rep     rep
	}{
		{
			name:    "create new short url",
			urlPath: "/",
			method:  http.MethodPost,
			body:    []byte("example.com"),
			want: want{
				code:        http.StatusCreated,
				response:    "http://localhost:8080/id",
				contentType: "text/plain",
			},
			rep: rep{},
		},
		{
			name:    "create new short url with empty body",
			urlPath: "/",
			method:  http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "body must have url\n",
				contentType: "text/plain; charset=utf-8",
			},
			rep: rep{},
		},
		{
			name:    "redirect with id",
			urlPath: "/id",
			method:  http.MethodGet,
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "<a href=\"/example.com\">Temporary Redirect</a>.\n\n",
				contentType: "text/html; charset=utf-8",
			},
			rep: rep{},
		},
		{
			name:    "redirect without id",
			urlPath: "/",
			method:  http.MethodGet,
			want: want{
				code:        http.StatusBadRequest,
				response:    "ID not provided\n",
				contentType: "text/plain; charset=utf-8",
			},
			rep: rep{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := bytes.NewReader(test.body)
			request := httptest.NewRequest(test.method, test.urlPath, reader)
			w := httptest.NewRecorder()

			handlers.URLHandler(&test.rep)(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
