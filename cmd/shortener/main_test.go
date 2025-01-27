package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	baseAddr = "http://localhosr:8000"
	testID   = "test-id"
)

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Store(id, value string) error {
	args := m.Called(id, value)
	return args.Error(0)
}

func (m *mockStorage) Get(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

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

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) (*http.Response, string) {
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

	longURLJSON := router.LongURL{URL: "http://example.com"}
	jsonLongURLBody, err := json.Marshal(longURLJSON)
	require.NoError(t, err)

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
			name:   "get '/{id}' with wrong id",
			url:    "/wrong-id",
			method: http.MethodGet,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "post '/api/shorten'",
			url:    "/api/shorten",
			method: http.MethodPost,
			body:   jsonLongURLBody,
			want: want{
				code: http.StatusCreated,
			},
		},
		{
			name:   "post '/api/shorten' empty body",
			url:    "/api/shorten",
			method: http.MethodPost,
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

func TestRepository_Initializing(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	var testTable = []struct {
		name    string
		options []repository.Option
		storage *mockStorage
	}{
		{
			name:    "without any options",
			storage: &mockStorage{},
		},
		{
			name:    "only with dump file",
			storage: &mockStorage{},
			options: []repository.Option{repository.AddDumpFile(file.Name())},
		},
		{
			name:    "only restore from dump file",
			storage: &mockStorage{},
			options: []repository.Option{repository.RestoreFromDump(file.Name())},
		},
		{
			name:    "with all options",
			storage: &mockStorage{},
			options: []repository.Option{
				repository.RestoreFromDump(file.Name()),
				repository.AddDumpFile(file.Name()),
			},
		},
	}

	for _, v := range testTable {
		t.Run(v.name, func(t *testing.T) {
			_, err := repository.NewRepository(v.storage)
			assert.NoError(t, err)
		})
	}
}

func TestRepository_Get(t *testing.T) {
	storage := &mockStorage{}
	storage.On("Get", "short").Return("long", nil)
	repo, _ := repository.NewRepository(storage)

	result, err := repo.Get("short")

	assert.NoError(t, err)
	assert.Equal(t, "long", result)
	storage.AssertCalled(t, "Get", "short")
}

func TestRepository_Store(t *testing.T) {
	storage := &mockStorage{}
	storage.On("Store", "short").Return("long", nil)
	repo, err := repository.NewRepository(storage)
	assert.NoError(t, err)
	storage.On("Store", "short", "long").Return(nil)
	err = repo.Store("short", "long")
	assert.NoError(t, err)
}

func TestRepository_StoreToFile(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	storage := &mockStorage{}
	repo, err := repository.NewRepository(storage, repository.AddDumpFile(file.Name()))
	assert.NoError(t, err)
	storage.On("Store", "short", "long").Return(nil)
	err = repo.Store("short", "long")
	assert.NoError(t, err)

	var storedRow row
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&storedRow)
	assert.NoError(t, err)
	assert.Equal(t, "short", storedRow.ShortURL)
	assert.Equal(t, "long", storedRow.LongURL)
}

func TestRepository_RestoreFromDump(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	rows := []row{
		{"short1", "long1"},
		{"short2", "long2"},
	}
	for _, r := range rows {
		data, _ := json.Marshal(r)
		file.Write(append(data, '\n'))
	}

	file.Close()

	storage := &mockStorage{}
	storage.On("Store", "short1", "long1").Return(nil)
	storage.On("Store", "short2", "long2").Return(nil)

	_, err = repository.NewRepository(storage, repository.RestoreFromDump(file.Name()))
	assert.NoError(t, err)

	storage.AssertCalled(t, "Store", "short1", "long1")
	storage.AssertCalled(t, "Store", "short2", "long2")
}

func TestRepository_Close(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	storage := &mockStorage{}
	repo, err := repository.NewRepository(storage, repository.AddDumpFile(file.Name()))
	assert.NoError(t, err)

	err = repo.Close()
	assert.NoError(t, err)
}

func TestRepository_StoreWithError(t *testing.T) {
	storage := &mockStorage{}
	storage.On("Store", "short", "long").Return(errors.New("storage error"))
	repo, _ := repository.NewRepository(storage)

	err := repo.Store("short", "long")

	assert.Error(t, err)
	assert.Equal(t, "storage error", err.Error())
}

func TestURLShortenerService(t *testing.T) {
	longValidURL := "https://validurl.com"
	longNOTValidURL := "NOT valid url.com"
	storage := &mockStorage{}
	storage.On("Store", mock.Anything, mock.Anything).Return(nil)
	storage.On("Get", testID).Return(longValidURL, nil)
	repo, err := repository.NewRepository(storage)
	assert.NoError(t, err)

	ser := service.NewURLShortener(repo, baseAddr)

	t.Run("Shorten valid url", func(t *testing.T) {
		shortURL, err := ser.ShortenURL(longValidURL)
		assert.NoError(t, err)
		assert.NotEqual(t, shortURL, longValidURL)
		assert.Contains(t, shortURL, baseAddr)
	})

	t.Run("Shorten NOT valid url", func(t *testing.T) {
		_, err := ser.ShortenURL(longNOTValidURL)
		assert.Error(t, err)
	})

	t.Run("Find by Alias(Shortened)", func(t *testing.T) {
		res, err := ser.FindByShortened(testID)
		assert.NoError(t, err)
		assert.Equal(t, longValidURL, res)
	})
}
