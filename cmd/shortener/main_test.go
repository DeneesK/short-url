package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	baseAddr = "http://localhosr:8000"
	testID   = "test-id"
	wrongID  = "wrong-id"
)

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	UserID   string `json:"user_id,omitempty"`
}

type ShortenerURLServiceMock struct {
	mock.Mock
}

func (m *ShortenerURLServiceMock) ShortenURL(ctx context.Context, value, userID string) (string, error) {
	args := m.Called(value)
	return args.String(0), args.Error(1)
}

func (m *ShortenerURLServiceMock) FindByShortened(ctx context.Context, id string) (dto.LongUrl, error) {
	args := m.Called(id)
	longURL := args.String(0)
	return dto.LongUrl{LongURL: longURL}, args.Error(1)
}

func (m *ShortenerURLServiceMock) FindByUserID(ctx context.Context, userID string) ([]dto.URL, error) {
	return nil, nil
}

func (m *ShortenerURLServiceMock) PingDB(ctx context.Context) error {
	return nil
}

func (m *ShortenerURLServiceMock) StoreBatchURL(ctx context.Context, batch []dto.OriginalURL, userID string) ([]dto.ShortedURL, error) {
	return nil, nil
}

func (m *ShortenerURLServiceMock) DeleteBatch(idx []string, userID string) {
}

type userServiceMock struct {
	mock.Mock
}

func (m *userServiceMock) Create(ctx context.Context) (string, error) {
	return "user_id:user_id", nil
}
func (m *userServiceMock) Verify(user string) bool {
	return true
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
	rep := &ShortenerURLServiceMock{}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	rep.On("ShortenURL", "http://example.com").Return(testID, nil)
	rep.On("FindByShortened", testID).Return("http://example.com", nil)
	rep.On("FindByShortened", wrongID).Return("", errors.New("id not found"))

	sugar := *logger.Sugar()
	userService := &userServiceMock{}
	r := router.NewRouter(rep, userService, &sugar)
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
	}{
		{
			name: "without any options",
		},
		{
			name:    "only with dump file",
			options: []repository.Option{repository.AddDumpFile(file.Name())},
		},
		{
			name:    "only restore from dump file",
			options: []repository.Option{repository.RestoreFromDump(file.Name())},
		},
		{
			options: []repository.Option{
				repository.RestoreFromDump(file.Name()),
				repository.AddDumpFile(file.Name()),
			},
		},
	}

	for _, v := range testTable {
		t.Run(v.name, func(t *testing.T) {
			_, err := repository.NewRepository(repository.StorageConfig{})
			assert.NoError(t, err)
		})
	}
}

func TestRepository_Store(t *testing.T) {
	repo, err := repository.NewRepository(repository.StorageConfig{})
	assert.NoError(t, err)
	_, err = repo.Store(context.TODO(), "id", "url", "user_id")
	assert.NoError(t, err)
}

func TestRepository_Get(t *testing.T) {
	repo, err := repository.NewRepository(repository.StorageConfig{MaxStorageSize: 100_000})
	assert.NoError(t, err)

	_, err = repo.Store(context.TODO(), "id", "url", "user_id")
	assert.NoError(t, err)

	result, err := repo.Get(context.TODO(), "id")

	assert.NoError(t, err)
	assert.Equal(t, "url", result.LongURL)
}

func TestRepository_StoreToFile(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	repo, err := repository.NewRepository(repository.StorageConfig{MaxStorageSize: 100_000}, repository.AddDumpFile(file.Name()))
	assert.NoError(t, err)
	_, err = repo.Store(context.TODO(), "short", "long", "user_id")
	assert.NoError(t, err)

	var storedRow row
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&storedRow)
	assert.NoError(t, err)
	assert.Equal(t, "short", storedRow.ShortURL)
	assert.Equal(t, "long", storedRow.LongURL)
	assert.Equal(t, "user_id", storedRow.UserID)
}

func TestRepository_RestoreFromDump(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	rows := []row{
		{"short1", "long1", ""},
		{"short2", "long2", ""},
	}
	for _, r := range rows {
		data, _ := json.Marshal(r)
		file.Write(append(data, '\n'))
	}

	file.Close()

	rep, err := repository.NewRepository(repository.StorageConfig{MaxStorageSize: 100_000}, repository.RestoreFromDump(file.Name()))
	assert.NoError(t, err)

	result, err := rep.Get(context.TODO(), "short1")
	assert.NoError(t, err)
	assert.Equal(t, "long1", result.LongURL)
}

func TestRepository_Close(t *testing.T) {
	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	repo, err := repository.NewRepository(repository.StorageConfig{MaxStorageSize: 100_000}, repository.AddDumpFile(file.Name()))
	assert.NoError(t, err)

	err = repo.Close(context.TODO())
	assert.NoError(t, err)
}

func TestURLShortenerService(t *testing.T) {
	longValidURL := "https://validurl.com"
	longValidURL2 := "https://validurl2.com"
	longNOTValidURL := "NOT valid url.com"
	repo, err := repository.NewRepository(repository.StorageConfig{MaxStorageSize: 100_000})
	assert.NoError(t, err)
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	ser := services.NewURLShortener(repo, baseAddr, logger.Sugar())

	t.Run("Shorten valid url", func(t *testing.T) {
		shortURL, err := ser.ShortenURL(context.TODO(), longValidURL, "user_id")
		assert.NoError(t, err)
		assert.NotEqual(t, shortURL, longValidURL)
		assert.Contains(t, shortURL, baseAddr)
	})

	t.Run("Shorten NOT valid url", func(t *testing.T) {
		_, err := ser.ShortenURL(context.TODO(), longNOTValidURL, "user_id")
		assert.Error(t, err)
	})

	t.Run("Shorten EXISTS valid long url", func(t *testing.T) {
		_, err := ser.ShortenURL(context.TODO(), longValidURL, "user_id")
		assert.ErrorIs(t, services.ErrLongURLAlreadyExists, err)
	})

	t.Run("Find by Alias(Shortened)", func(t *testing.T) {
		shortURL, err := ser.ShortenURL(context.TODO(), longValidURL2, "user_id")
		assert.NoError(t, err)
		id := (strings.Split(shortURL, baseAddr+"/"))[1]
		res, err := ser.FindByShortened(context.TODO(), id)
		assert.NoError(t, err)
		assert.Equal(t, longValidURL2, res.LongURL)
	})
}
