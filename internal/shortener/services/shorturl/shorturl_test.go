package shorturl

import (
	"context"
	"fmt"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type saverMock struct {
	SaveFn  func(context.Context, models.ShortenedURL) error
	BatchFn func(ctx context.Context, urls []models.ShortenedURL) error
}

func (m *saverMock) Save(ctx context.Context, data models.ShortenedURL) error {
	if m != nil && m.SaveFn != nil {
		return m.SaveFn(ctx, data)
	}
	return fmt.Errorf("unable to save")
}

func (m *saverMock) Batch(ctx context.Context, urls []models.ShortenedURL) error {
	if m != nil && m.BatchFn != nil {
		return m.BatchFn(ctx, urls)
	}
	return fmt.Errorf("unable to batch")
}

func TestShortener_Short(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
	}
	tests := []struct {
		url     url
		err     error
		saver   saverMock
		baseURL string
		name    string
	}{
		{
			name:    "Short http://example.com/query, no error",
			baseURL: "http://localhost:8080",
			url: url{
				userID: "1",
				corrID: "1",
				raw:    "http://example.com/query",
			},
			saver: saverMock{
				SaveFn: func(_ context.Context, data models.ShortenedURL) error {
					return nil
				},
			},
		},
		{
			name:    "Short http://192.168.1.1/query, no error",
			baseURL: "http://localhost:8080",
			url: url{
				userID: "1",
				corrID: "1",
				raw:    "http://192.168.1.1/query",
			},
			saver: saverMock{
				SaveFn: func(_ context.Context, data models.ShortenedURL) error {
					return nil
				},
			},
		},
		{
			name:    "Short http://192.168.1.1/query, error expected",
			baseURL: "http://localhost:8080",
			url: url{
				userID: "1",
				corrID: "1",
				raw:    "http://192.168.1.1/query",
			},
			saver: saverMock{
				SaveFn: func(_ context.Context, data models.ShortenedURL) error {
					return shortenedurl.ErrUniqueViolation
				},
			},
			err: ErrUniqueViolation,
		},
		{
			name:    "Internal error expected",
			baseURL: "http://localhost:8080",
			url: url{
				userID: "1",
				corrID: "1",
				raw:    "http://192.168.1.1/query",
			},
			saver: saverMock{
				SaveFn: func(_ context.Context, data models.ShortenedURL) error {
					return fmt.Errorf("internal error")
				},
			},
			err: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shortener{
				baseURL: tt.baseURL,
				saver:   &tt.saver,
			}

			url := models.NewURL(tt.url.userID, tt.url.corrID, tt.url.raw)

			got, err := service.Short(context.TODO(), url)
			if tt.err == nil {
				require.NoError(t, err)
				assert.NotEmpty(t, got)
				return
			}
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestShortener_Batch(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
	}
	tests := []struct {
		saver   saverMock
		urls    []url
		err     error
		name    string
		baseURL string
	}{
		{
			name:    "Batch all urls, no error",
			baseURL: "http://localhost:8080",
			urls: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://example.com/query1",
				},
				{
					userID: "1",
					corrID: "2",
					raw:    "http://example.com/query2",
				},
				{
					userID: "1",
					corrID: "3",
					raw:    "http://example.com/query3",
				},
				{
					userID: "1",
					corrID: "4",
					raw:    "http://example.com/query4",
				},
				{
					userID: "1",
					corrID: "5",
					raw:    "http://example.com/query5",
				},
			},
			saver: saverMock{
				BatchFn: func(ctx context.Context, su []models.ShortenedURL) error {
					return nil
				},
			},
		},
		{
			name:    "Batch all urls, empty userID",
			baseURL: "http://localhost:8080",
			urls: []url{
				{
					corrID: "1",
					raw:    "http://example.com/query1",
				},
				{
					userID: "1",
					corrID: "2",
					raw:    "http://example.com/query2",
				},
				{
					userID: "1",
					corrID: "3",
					raw:    "http://example.com/query3",
				},
				{
					userID: "1",
					corrID: "4",
					raw:    "http://example.com/query4",
				},
				{
					userID: "1",
					corrID: "5",
					raw:    "http://example.com/query5",
				},
			},
			saver: saverMock{},
			err:   ErrInvalidCreation,
		},
		{
			name:    "Internal batch error",
			baseURL: "http://localhost:8080",
			urls: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://example.com/query1",
				},
				{
					userID: "1",
					corrID: "2",
					raw:    "http://example.com/query2",
				},
				{
					userID: "1",
					corrID: "3",
					raw:    "http://example.com/query3",
				},
				{
					userID: "1",
					corrID: "4",
					raw:    "http://example.com/query4",
				},
				{
					userID: "1",
					corrID: "5",
					raw:    "http://example.com/query5",
				},
			},
			saver: saverMock{},
			err:   fmt.Errorf("unable to batch"),
		},
		{
			name:    "Empty batch",
			baseURL: "http://localhost:8080",
			err:     ErrEmptyBatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shortener{
				baseURL: "http://localhost:8080",
				saver:   &tt.saver,
			}

			batchURLs := make([]models.URL, len(tt.urls))
			for i, v := range tt.urls {
				batchURLs[i] = models.URL{
					UserID: v.userID,
					CorrID: v.corrID,
					Raw:    v.raw,
				}
			}

			batched, err := service.Batch(context.TODO(), batchURLs)
			if tt.err == nil {
				require.NoError(t, err)

				for i, v := range batched {
					assert.Equal(t, tt.urls[i].userID, v.UserID)
					assert.Equal(t, tt.urls[i].corrID, v.CorrID)
					assert.Equal(t, tt.urls[i].raw, v.Raw)
				}
				return
			}
			assert.Equal(t, tt.err.Error(), err.Error())
		})
	}
}
