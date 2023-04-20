package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/services/shorturl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shortenerMock struct {
	ShortFn func(context.Context, models.URL) (string, error)
}

func (m *shortenerMock) Short(ctx context.Context, url models.URL) (string, error) {
	if m != nil && m.ShortFn != nil {
		return m.ShortFn(ctx, url)
	}
	return "", fmt.Errorf("unable to short URL")
}

type getterByURLMock struct {
	GetByURLFn func(context.Context, string) (models.ShortenedURL, error)
}

func (m *getterByURLMock) GetByURL(ctx context.Context, raw string) (models.ShortenedURL, error) {
	if m != nil && m.GetByURLFn != nil {
		return m.GetByURLFn(ctx, raw)
	}
	return models.ShortenedURL{}, fmt.Errorf("unable to short URL")
}

func TestShort(t *testing.T) {
	type services struct {
		shortsrv shortener
		getter   getterByURL
	}
	type request struct {
		body string
	}
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		serv services
		req  request
		name string
		want want
	}{
		{
			name: "New URL, status code: Created",
			req: request{
				body: "https://demo.com",
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "https://demo.com/tmp_slug", nil
					},
				},
			},
		},
		{
			name: "Existed URL, status code: Conflict",
			req: request{
				body: "https//go.dev",
			},
			want: want{
				code:        http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrUniqueViolation
					},
				},
				getter: &getterByURLMock{
					GetByURLFn: func(_ context.Context, s string) (models.ShortenedURL, error) {
						return models.NewShortenedURL("1", "1", "https//go-dev",
							"tmp_slug", "http//localhost:8080/tmp_slug"), nil
					},
				},
			},
		},
		{
			name: "Empty request, status code: BadRequest",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrInvalidCreation
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.POST("/", authWrap(short(tt.serv.shortsrv, tt.serv.getter)))

			w := httptest.NewRecorder()
			// Prepare request.
			req := textReq(t, "/", http.MethodPost, bytes.NewBufferString(tt.req.body))
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want.code, resp.StatusCode,
				"Expected status code: %d, got %d", tt.want.code, resp.StatusCode)
			assert.EqualValues(t, tt.want.contentType, resp.Header.Get("Content-Type"),
				"Expected Content-Type Header: %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.want.code == http.StatusCreated || tt.want.code == http.StatusConflict {
				require.NotEmpty(t, respBody)
			}
		})
	}
}

func TestShorten(t *testing.T) {
	type services struct {
		shortsrv shortener
		provider getterByURL
	}
	type request struct {
		body shortURLRequest
	}
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		serv services
		req  request
		name string
		want want
	}{
		{
			name: "New URL, status code: Created",
			req: request{
				body: shortURLRequest{
					URL: "https://demo.com",
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "shorted URL", nil
					},
				},
			},
		},
		{
			name: "Existed URL, status code: Conflict",
			req: request{
				body: shortURLRequest{
					URL: "https://demo.com",
				},
			},
			want: want{
				code:        http.StatusConflict,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrUniqueViolation
					},
				},
				provider: &getterByURLMock{
					GetByURLFn: func(_ context.Context, s string) (models.ShortenedURL, error) {
						return models.NewShortenedURL("", "1", "https//go-dev",
							"tmp_slug", "http://localhost:8080/tmp_slug"), nil
					},
				},
			},
		},
		{
			name: "Empty request, status code: BadRequest",
			req: request{
				body: shortURLRequest{
					URL: "",
				},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrInvalidCreation
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.POST("/api/shorten", authWrap(shorten(tt.serv.shortsrv, tt.serv.provider)))

			w := httptest.NewRecorder()
			// Prepare the request.
			req := shortURLJSONReq(t, "/api/shorten", http.MethodPost, tt.req.body)
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.EqualValues(t, tt.want.code, resp.StatusCode,
				"Expected status code: %d, got %d", tt.want.code, resp.StatusCode)
			assert.EqualValues(t, tt.want.contentType, resp.Header.Get("Content-Type"),
				"Expected Content-Type Header: %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))

			if tt.want.code == http.StatusCreated || tt.want.code == http.StatusConflict {
				resBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.NotEmpty(t, resBody)
			}
		})
	}
}

// jsonReq prepares http.Request with a JSON body.
func shortURLJSONReq(t *testing.T, api, method string, reqBody shortURLRequest) *http.Request {
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(method, api, bytes.NewBuffer(data))
	require.NoError(t, err)

	return req
}
