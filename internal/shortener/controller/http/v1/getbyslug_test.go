package v1

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type getterBySlugMock struct {
	GetBySlugFn func(context.Context, string) (models.ShortenedURL, error)
}

func (m *getterBySlugMock) GetBySlug(ctx context.Context, slug string) (models.ShortenedURL, error) {
	if m != nil && m.GetBySlugFn != nil {
		return m.GetBySlugFn(ctx, slug)
	}
	return models.ShortenedURL{}, fmt.Errorf("unable to get an URL using slug")
}

func TestGetBySlugRoute_GetBySlug(t *testing.T) {
	type services struct {
		getter getterBySlug
	}
	type request struct {
		api    string
		method string
	}
	type want struct {
		contentType string
		data        string
		code        int
	}
	tests := []struct {
		name              string
		req               request
		serv              services
		want              want
		locationHeaderSet bool
	}{
		{
			name: "URL by 112Sd exists, status code: TemporaryRedirect",
			req: request{
				api:    "/112Sd",
				method: http.MethodGet,
			},
			want: want{
				data:        "http://example.com/query_1",
				code:        http.StatusTemporaryRedirect,
				contentType: "text/plain; charset=utf-8",
			},
			serv: services{
				getter: &getterBySlugMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						return models.NewShortenedURL("", "1", "http://example.com/query_1",
							"tmp_slug", "http://localhost:8080/tmp_slug"), nil
					},
				},
			},
			locationHeaderSet: true,
		},
		{
			name: "URL by 112Sd was deleted, status code: Gone",
			req: request{
				api:    "/112Sd",
				method: http.MethodGet,
			},
			want: want{
				code:        http.StatusGone,
				contentType: "text/plain; charset=utf-8",
			},
			serv: services{
				getter: &getterBySlugMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						url := models.NewShortenedURL("1", "1", "http://example.com/query_1",
							"tmp_slug", "http://localhost:8080/tmp_slug")
						url.IsDeleted = true
						return url, nil
					},
				},
			},
		},
		{
			name: "URL doesn't exist, status code: NotFound",
			req: request{
				api:    "/any_slug",
				method: http.MethodGet,
			},
			want: want{
				code: http.StatusNotFound,
			},
			serv: services{
				getter: &getterBySlugMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						return models.ShortenedURL{}, nil
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.GET("/:slug", getBySlug(tt.serv.getter))

			w := httptest.NewRecorder()
			// Prepare the request.
			req, err := http.NewRequest("GET", tt.req.api, nil)
			require.NoError(t, err)

			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Discard response body.
			_, err = io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)

			require.Equal(t, tt.want.code, resp.StatusCode,
				"Expected response status code: %d, got %d", tt.want.code, resp.StatusCode)

			if tt.locationHeaderSet {
				assert.Equal(t, tt.want.data, resp.Header.Get("Location"),
					"Expected Location Header: %s, got %s", tt.want.data, resp.Header.Get("Location"))
			}
		})
	}
}
