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

type batcherMock struct {
	BatchFn func(context.Context, []models.URL) ([]models.ShortenedURL, error)
}

func (m *batcherMock) Batch(ctx context.Context, urls []models.URL) ([]models.ShortenedURL, error) {
	if m != nil && m.BatchFn != nil {
		return m.BatchFn(ctx, urls)
	}
	return nil, fmt.Errorf("unable to batch")
}

func TestBatchURLRoute_Batch(t *testing.T) {
	type services struct {
		batcher batcher
	}
	type request struct {
		body []batchURLsRequest
	}
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		want want
		serv services
		req  request
	}{
		{
			name: "List URLs, status code: Ok",
			req: request{
				body: []batchURLsRequest{
					{
						CorrID: "0001",
						RawURL: "http://example.com/query_1",
					},
					{
						CorrID: "0002",
						RawURL: "http://example.com/query_2",
					},
					{
						CorrID: "0003",
						RawURL: "http://example.com/query_3",
					},
					{
						CorrID: "0004",
						RawURL: "http://example.com/query_4",
					},
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				batcher: &batcherMock{
					BatchFn: func(ctx context.Context, u []models.URL) ([]models.ShortenedURL, error) {
						records := []models.ShortenedURL{
							{
								CorrID: "0001",
								Raw:    "http://example.com/query_1",
							},
							{
								CorrID: "0002",
								Raw:    "http://example.com/query_2",
							},
							{
								CorrID: "0003",
								Raw:    "http://example.com/query_3",
							},
							{
								CorrID: "0004",
								Raw:    "http://example.com/query_4",
							},
						}
						return records, nil
					},
				},
			},
		},
		{
			name: "No URLs for batching, status code: BadRequest",
			req: request{
				body: []batchURLsRequest{},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				batcher: &batcherMock{
					BatchFn: func(ctx context.Context, u []models.URL) ([]models.ShortenedURL, error) {
						return nil, shorturl.ErrEmptyBatch
					},
				},
			},
		},
		{
			name: "Invalid URLs for batching, status code: BadRequest",
			req: request{
				body: []batchURLsRequest{
					{
						CorrID: "0001",
						RawURL: "httpa://example.com/query_1",
					},
					{
						CorrID: "0002",
						RawURL: "http//example.com/query_2",
					},
					{
						CorrID: "0003",
						RawURL: "http:example.com/query_3",
					},
				},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
			serv: services{
				batcher: &batcherMock{
					BatchFn: func(ctx context.Context, u []models.URL) ([]models.ShortenedURL, error) {
						return nil, shorturl.ErrInvalidCreation
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.POST("/api/shorten/batch", authWrap(batchURLs(tt.serv.batcher)))

			w := httptest.NewRecorder()
			// Prepare the request.
			req := batchURLsJSONReq(t, "/api/shorten/batch", http.MethodPost, tt.req.body)
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want.code, resp.StatusCode)

			// Get response.
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if len(respBody) > 0 {
				var batchedURLs []batchURLsResponse
				err = json.Unmarshal(respBody, &batchedURLs)
				require.NoError(t, err)

				assert.EqualValues(t, len(tt.req.body), len(batchedURLs))
				assert.EqualValues(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}

		})
	}
}

// jsonReq prepares http.Request with a JSON body.
func batchURLsJSONReq(t *testing.T, api, method string, reqBody []batchURLsRequest) *http.Request {
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(method, api, bytes.NewBuffer(data))
	require.NoError(t, err)

	return req
}
