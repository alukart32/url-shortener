package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

type collectorMock struct {
	CollectByUserFn func(context.Context, string) ([]models.ShortenedURL, error)
}

func (m *collectorMock) CollectByUser(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	if m != nil && m.CollectByUserFn != nil {
		return m.CollectByUserFn(ctx, userID)
	}
	return nil, fmt.Errorf("unable to collect URLs.")
}

func TestCollectURLs(t *testing.T) {
	type services struct {
		collector collector
	}
	type want struct {
		contentType string
		data        []collectURLsResponse
		code        int
	}
	tests := []struct {
		name string
		serv services
		want want
	}{
		{
			name: "List URLs, status code: Ok",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json; charset=utf-8",
				data: []collectURLsResponse{
					{
						OriginalURL: "http://example.com/query_1",
					},
					{
						OriginalURL: "http://example.com/query_2",
					},
				},
			},
			serv: services{
				collector: &collectorMock{
					CollectByUserFn: func(ctx context.Context, s string) (
						[]models.ShortenedURL, error) {
						records := []models.ShortenedURL{
							{
								Raw: "http://example.com/query_1",
							},
							{
								Raw: "http://example.com/query_2",
							},
						}
						return records, nil
					},
				},
			},
		},
		{
			name: "No URLs, status code: NoContent",
			want: want{
				code:        http.StatusNoContent,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				collector: &collectorMock{
					CollectByUserFn: func(ctx context.Context, s string) (
						[]models.ShortenedURL, error) {
						return []models.ShortenedURL{}, nil
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.GET("/api/user/urls", authWrap(collectURLs(tt.serv.collector)))

			w := httptest.NewRecorder()
			// Prepare request.
			req := textReq(t, "/api/user/urls", http.MethodGet, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want.code, resp.StatusCode)

			// Get response.
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if len(respBody) > 0 {
				var collectedURLs []collectURLsResponse
				err = json.Unmarshal(respBody, &collectedURLs)
				require.NoError(t, err)

				assert.EqualValues(t, len(tt.want.data), len(collectedURLs))
				assert.EqualValues(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}
