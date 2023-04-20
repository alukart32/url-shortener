package memstorage

import (
	"context"
	"fmt"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_Save(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		args       url
		existedURL url
		name       string
		baseURL    string
		urlExistes bool
	}{
		{
			name:    "New URL",
			baseURL: "http://127.0.0.1/",
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
		},
		{
			name:       "Save existed URL",
			baseURL:    "http://127.0.0.1/",
			urlExistes: true,
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug2",
			},
			existedURL: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			if tt.urlExistes {
				existedURL := models.NewShortenedURL(
					tt.existedURL.userID,
					tt.existedURL.corrID,
					tt.existedURL.raw,
					tt.existedURL.slug,
					tt.baseURL+tt.existedURL.slug,
				)
				require.NoError(t, storage.Save(context.TODO(), existedURL))
			}

			shortenedURL := models.NewShortenedURL(
				tt.args.userID,
				tt.args.corrID,
				tt.args.raw,
				tt.args.slug,
				tt.baseURL+tt.args.slug,
			)

			err := storage.Save(context.TODO(), shortenedURL)
			if !tt.urlExistes {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
		})
	}
}

func BenchmarkMemStorage_Save(b *testing.B) {
	n := 100

	userID := "1"
	shortenedURLs := make([]models.ShortenedURL, n)
	for i := 0; i < n; i++ {
		slug := fmt.Sprintf("slug_%d", i)

		if i%10 == 0 {
			userID = fmt.Sprintf("%d", i)
		}
		v := models.NewShortenedURL(
			userID,
			fmt.Sprintf("%d", i),
			fmt.Sprintf("http://demo.com/%v", i),
			slug,
			"http://127.0.0.1:8080/"+slug,
		)

		shortenedURLs[i] = v
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		storage := MemStorage()
		ctx := context.TODO()
		b.StartTimer()

		for _, v := range shortenedURLs {
			if err := storage.Save(ctx, v); err != nil {
				b.Fail()
			}
		}
	}
}

func TestMemStorage_GetByURL(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		args       url
		existedURL url
		name       string
		baseURL    string
		urlExists  bool
	}{
		{
			name:      "URL exists",
			baseURL:   "http://127.0.0.1/",
			urlExists: true,
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
			existedURL: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
		},
		{
			name:    "URL not found",
			baseURL: "http://127.0.0.1/",
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
			existedURL: url{
				userID: "1",
				corrID: "2",
				raw:    "http://demo2.com",
				slug:   "slug2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			if tt.urlExists {
				existedURL := models.NewShortenedURL(
					tt.existedURL.userID,
					tt.existedURL.corrID,
					tt.existedURL.raw,
					tt.existedURL.slug,
					tt.baseURL+tt.existedURL.slug,
				)

				require.NoError(t, storage.Save(context.TODO(), existedURL))
			}

			shortenedURL, err := storage.GetByURL(context.TODO(), tt.args.raw)
			require.NoError(t, err)

			if tt.urlExists {
				assert.Equal(t, tt.args.userID, shortenedURL.UserID)
				assert.Equal(t, tt.args.raw, shortenedURL.Raw)
				return
			}

		})
	}
}

func TestMemStorage_GetBySlug(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		args    url
		name    string
		baseURL string
		slug    string
		found   bool
	}{
		{
			name:    "URL exists",
			baseURL: "http://127.0.0.1/",
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
			slug:  "slug1",
			found: true,
		},
		{
			name: "URL not found",
			args: url{
				userID: "1",
				corrID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
			slug: "no",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			shortenedURL := models.NewShortenedURL(
				tt.args.userID,
				tt.args.corrID,
				tt.args.raw,
				tt.args.slug,
				tt.baseURL+tt.args.slug,
			)

			require.NoError(t, storage.Save(context.TODO(), shortenedURL))

			got, err := storage.GetBySlug(context.TODO(), tt.slug)
			require.NoError(t, err)

			if tt.found {
				assert.Equal(t, shortenedURL.UserID, got.UserID)
				assert.Equal(t, shortenedURL.Raw, got.Raw)
				return
			}
		})
	}
}

func TestMemStorage_Collect(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		name           string
		baseURL        string
		collectForUser string
		existedURLs    []url
		wantErr        bool
	}{
		{
			name:           "URLs found",
			baseURL:        "http://127.0.0.1/",
			collectForUser: "1",
			existedURLs: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com",
					slug:   "slug1",
				},
			},
		},
		{
			name:           "URLs not found",
			baseURL:        "http://127.0.0.1/",
			collectForUser: "1",
			existedURLs: []url{
				{
					userID: "2",
					corrID: "2",
					raw:    "http://demo.com",
					slug:   "slug2",
				},
			},
			wantErr: true,
		},
		{
			name:    "UserID is empty",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			for _, v := range tt.existedURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)

				require.NoError(t, storage.Save(context.TODO(), shortenedURL))
			}

			collected, err := storage.CollectByUser(context.TODO(), tt.collectForUser)
			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, 1, len(collected))
				for i, v := range tt.existedURLs {
					assert.Equal(t, v.userID, collected[i].UserID)
					assert.Equal(t, v.raw, collected[i].Raw)
				}
				return
			}
			assert.True(t, len(collected) == 0)
		})
	}
}

func TestMemStorage_Batch(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		name    string
		baseURL string
		args    []url
	}{
		{
			name:    "Batch",
			baseURL: "http://127.0.0.1",
			args: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com/1",
					slug:   "slug1",
				},
				{
					userID: "2",
					corrID: "2",
					raw:    "http://demo.com/2",
					slug:   "slug2",
				},
				{
					userID: "1",
					corrID: "3",
					raw:    "http://demo.com/3",
					slug:   "slug3",
				},
				{
					userID: "3",
					corrID: "4",
					raw:    "http://demo.com/4",
					slug:   "slug4",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			shortenedURLs := make([]models.ShortenedURL, len(tt.args))
			for i, v := range tt.args {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)
				shortenedURLs[i] = shortenedURL
			}

			err := storage.Batch(context.TODO(), shortenedURLs)
			assert.NoError(t, err)
		})
	}
}

func TestMemStorage_Delete(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		name         string
		baseURL      string
		targetUserID string
		args         []url
	}{
		{
			name:         "Delete",
			baseURL:      "http://127.0.0.1/",
			targetUserID: "1",
			args: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com/1",
					slug:   "slug1",
				},
				{
					userID: "1",
					corrID: "2",
					raw:    "http://demo.com/2",
					slug:   "slug2",
				},
				{
					userID: "1",
					corrID: "3",
					raw:    "http://demo.com/3",
					slug:   "slug3",
				},
				{
					userID: "1",
					corrID: "4",
					raw:    "http://demo.com/4",
					slug:   "slug4",
				},
				{
					userID: "1",
					corrID: "5",
					raw:    "http://demo.com/5",
					slug:   "slug5",
				},
				{
					userID: "2",
					corrID: "1",
					raw:    "http://test.com/1",
					slug:   "slug6",
				},
				{
					userID: "2",
					corrID: "2",
					raw:    "http://test.com/2",
					slug:   "slug7",
				},
				{
					userID: "2",
					corrID: "3",
					raw:    "http://test.com/3",
					slug:   "slug8",
				},
				{
					userID: "2",
					corrID: "4",
					raw:    "http://test.com/4",
					slug:   "slug9",
				},
				{
					userID: "2",
					corrID: "5",
					raw:    "http://test.com/5",
					slug:   "slug10",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			var slugs []string
			for _, v := range tt.args {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)
				if shortenedURL.UserID == tt.targetUserID {
					slugs = append(slugs, shortenedURL.Slug)
				}

				err := storage.Save(context.TODO(), shortenedURL)
				require.NoError(t, err)
			}

			err := storage.Delete(tt.targetUserID, slugs)
			require.NoError(t, err)

			for _, slug := range slugs {
				if v, ok := storage.data[slug]; ok && !v.IsDeleted {
					t.Errorf("url with slug: %v not deleted", slug)
				}
			}
		})
	}
}

func BenchmarkMemStorage_Delete(b *testing.B) {
	n := 100

	userID := "1"
	shortenedURLs := make([]models.ShortenedURL, n)
	for i := 0; i < n; i++ {
		slug := fmt.Sprintf("slug_%d", i)

		if i%10 == 0 {
			userID = fmt.Sprintf("%d", i)
		}
		v := models.NewShortenedURL(
			userID,
			fmt.Sprintf("%d", i),
			fmt.Sprintf("http://demo.com/%v", i),
			slug,
			"http://127.0.0.1:8080/"+slug,
		)

		shortenedURLs[i] = v
	}

	slugsCount := 30
	slugs := make([]string, slugsCount)
	for i := 0; i < slugsCount; i++ {
		slugs[i] = shortenedURLs[i+2].Slug
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		storage := MemStorage()
		b.StartTimer()

		if err := storage.Delete(userID, slugs); err != nil {
			b.Fail()
		}
	}
}

func TestMemStorage_Stat(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		existedURLs []url
		want        models.Stat
		name        string
		baseURL     string
	}{
		{
			name:    "Stat",
			baseURL: "http://127.0.0.1/",
			want: models.Stat{
				URLs:  10,
				Users: 7,
			},
			existedURLs: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com",
					slug:   "slug1",
				},
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo2.com",
					slug:   "slug2",
				},
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo3.com",
					slug:   "slug3",
				},
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo4.com",
					slug:   "slug4",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo5.com",
					slug:   "slug5",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo6.com",
					slug:   "slug6",
				},
				{
					userID: "4",
					corrID: "1",
					raw:    "http://demo7.com",
					slug:   "slug7",
				},
				{
					userID: "5",
					corrID: "1",
					raw:    "http://demo8.com",
					slug:   "slug8",
				},
				{
					userID: "6",
					corrID: "1",
					raw:    "http://demo9.com",
					slug:   "slug9",
				},
				{
					userID: "7",
					corrID: "1",
					raw:    "http://demo10.com",
					slug:   "slug10",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage()

			for _, v := range tt.existedURLs {
				existedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)

				require.NoError(t, storage.Save(context.TODO(), existedURL))
			}

			got, err := storage.Stat(context.TODO())
			require.NoError(t, err)

			assert.EqualValues(t, tt.want.URLs, got.URLs)
			assert.EqualValues(t, tt.want.Users, got.Users)
		})
	}
}
