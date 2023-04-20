package shorturl

import (
	"fmt"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {
	tests := []struct {
		url     models.URL
		err     error
		name    string
		baseURL string
	}{
		{
			name:    "Short valid URL",
			url:     models.NewURL("1", "1", "http://demo.com"),
			baseURL: "http://localhost:8080",
		},
		{
			name:    "Short URL, empty userID",
			url:     models.NewURL("", "1", "http://demo.com"),
			err:     fmt.Errorf("empty userID"),
			baseURL: "http://localhost:8080",
		},
		{
			name:    "Short URL, empty URI",
			url:     models.NewURL("1", "1", ""),
			err:     fmt.Errorf("empty URI"),
			baseURL: "http://localhost:8080",
		},
		{
			name:    "Short URL, invalid URI",
			url:     models.NewURL("1", "1", "httpdemo.com"),
			err:     fmt.Errorf("failed to parse URI"),
			baseURL: "http://localhost:8080",
		},
		{
			name: "Short URL, empty baseURL",
			url:  models.NewURL("1", "1", "http://demo.com"),
			err:  fmt.Errorf("empty baseURL"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := shortenURL(
				tt.url.UserID,
				tt.url.CorrID,
				tt.url.Raw,
				tt.baseURL,
			)

			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
				return
			}
		})
	}
}

func TestShortenedURL(t *testing.T) {
	s1, err := shortenURL("1", "1", "http://demo.com", "http://localhost:8080")
	require.NoError(t, err)

	s2, err := shortenURL("1", "1", "http://demo.com", "http://localhost:8080")
	require.NoError(t, err)

	s3, err := shortenURL("3", "1", "http://demo3.com", "http://localhost:8080")
	require.NoError(t, err)

	assert.False(t, s1.Equals(s2))
	assert.NotEqualValues(t, s1.String(), s2.String())

	s2.SetDeleted()
	assert.False(t, s1.Equals(s2))

	assert.False(t, s3.Empty())
}
