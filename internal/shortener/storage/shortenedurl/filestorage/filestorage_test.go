package filestorage

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_Save(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		args    url
		err     error
		name    string
		baseURL string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, close, err := newFileStorage("test_save")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			v := models.NewShortenedURL(
				tt.args.userID,
				tt.args.corrID,
				tt.args.raw,
				tt.args.slug,
				tt.baseURL+tt.args.slug,
			)

			err = r.Save(context.TODO(), v)
			if tt.err == nil {
				require.NoError(t, err)
				return
			}
		})
	}
}

func TestFileStorage_GetBySlug(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		existedURLs []url
		wantURL     url
		name        string
		baseURL     string
		slug        string
		notFound    bool
	}{
		{
			name:     "URL exists",
			baseURL:  "http://127.0.0.1/",
			notFound: false,
			wantURL: url{
				userID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
			existedURLs: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com",
					slug:   "slug1",
				},
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo2.com",
					slug:   "slug2",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo3.com",
					slug:   "slug3",
				},
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo4.com",
					slug:   "slug4",
				},
			},
		},
		{
			name:        "URL not found",
			existedURLs: nil,
			slug:        "112SD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, close, err := newFileStorage("test_getbyslug")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			var slug string

			for _, v := range tt.existedURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)

				err = r.Save(context.TODO(), shortenedURL)
				require.NoError(t, err)
			}

			// Set unreachable slug.
			if tt.notFound {
				slug = tt.slug
			}

			got, err := r.GetBySlug(context.TODO(), slug)
			require.NoError(t, err)
			if tt.notFound {
				assert.Equal(t, tt.wantURL.userID, got.UserID)
				assert.Equal(t, tt.wantURL.raw, got.Raw)
				assert.Equal(t, tt.wantURL.slug, got.Slug)
				return
			}
		})
	}
}

func BenchmarkFileStorage_GetBySlug(b *testing.B) {
	rand.NewSource(time.Now().Unix())

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

	repo, close, err := newFileStorage("bench_getbyslug")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := close(); err != nil {
			b.Fatal(err)
		}
	}()

	for _, v := range shortenedURLs {
		if err := repo.Save(context.TODO(), v); err != nil {
			b.Fail()
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		slug := shortenedURLs[rand.Intn(n)].Slug
		b.StartTimer()

		_, err = repo.GetBySlug(context.TODO(), slug)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileStorage_GetByURL(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		wantURL     url
		name        string
		baseURL     string
		existedURLs []url
		notFound    bool
	}{
		{
			name:    "URL exists",
			baseURL: "http://127.0.0.1/",
			wantURL: url{
				userID: "2",
				raw:    "http://demo.com/query_2",
				slug:   "slug2",
			},
			existedURLs: []url{
				{
					userID: "1",
					corrID: "1",
					raw:    "http://demo.com",
					slug:   "slug1",
				},
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo.com/query_2",
					slug:   "slug2",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo.com/query_3",
					slug:   "slug3",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo.com/query_4",
					slug:   "slug4",
				},
				{
					userID: "3",
					corrID: "1",
					raw:    "http://demo.com/query_5",
					slug:   "slug5",
				},
			},
		},
		{
			name:        "URL not found",
			baseURL:     "http://127.0.0.1/",
			existedURLs: nil,
			notFound:    true,
			wantURL: url{
				userID: "1",
				raw:    "http://demo.com",
				slug:   "slug1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, close, err := newFileStorage("test_getbyurl")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			for _, v := range tt.existedURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)
				err = r.Save(context.TODO(), shortenedURL)
				require.NoError(t, err)
			}

			got, err := r.GetByURL(context.TODO(), tt.wantURL.raw)
			require.NoError(t, err)
			if !tt.notFound {
				assert.Equal(t, tt.wantURL.userID, got.UserID)
				assert.Equal(t, tt.wantURL.raw, got.Raw)
				assert.Equal(t, tt.wantURL.slug, got.Slug)
				return
			}
		})
	}
}

func BenchmarkFileRepo_GetByURL(b *testing.B) {
	rand.NewSource(time.Now().Unix())

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

	repo, close, err := newFileStorage("bench_getbyurl")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := close(); err != nil {
			b.Fatal(err)
		}
	}()

	for _, v := range shortenedURLs {
		if err := repo.Save(context.TODO(), v); err != nil {
			b.Fail()
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		url := shortenedURLs[rand.Intn(n)].Raw
		b.StartTimer()

		_, err = repo.GetByURL(context.TODO(), url)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileRepo_Collect(t *testing.T) {
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
		wantURLs       []url
		existedURLs    []url
		noRecords      bool
	}{
		{
			name:           "URLs found",
			baseURL:        "http://127.0.0.1/",
			collectForUser: "1",
			wantURLs: []url{
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
					slug:   "slug1",
				},
			},
			existedURLs: []url{
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo2.com/",
					slug:   "slug21",
				},
				{
					userID: "3",
					corrID: "2",
					raw:    "http://demo3.com/",
					slug:   "slug32",
				},
				{
					userID: "4",
					corrID: "3",
					raw:    "http://demo4.com/",
					slug:   "slug14",
				},
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
					slug:   "slug1",
				},
			},
		},
		{
			name:           "URLs not found",
			baseURL:        "http://127.0.0.1/",
			collectForUser: "9999",
			noRecords:      true,
			existedURLs: []url{
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo2.com/",
					slug:   "slug21",
				},
				{
					userID: "3",
					corrID: "2",
					raw:    "http://demo3.com/",
					slug:   "slug32",
				},
				{
					userID: "4",
					corrID: "3",
					raw:    "http://demo4.com/",
					slug:   "slug14",
				},
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
					slug:   "slug1",
				},
			},
		},
		{
			name:      "UserID is empty",
			baseURL:   "http://127.0.0.1/",
			noRecords: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, close, err := newFileStorage("test_collect")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			for _, v := range tt.existedURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)

				err = r.Save(context.TODO(), shortenedURL)
				require.NoError(t, err)
			}

			records, err := r.CollectByUser(context.TODO(), tt.collectForUser)
			if !tt.noRecords {
				require.NoError(t, err)
				assert.Equal(t, len(tt.wantURLs), len(records))

				for i, v := range tt.wantURLs {
					assert.Equal(t, v.userID, records[i].UserID)
					assert.Equal(t, v.corrID, records[i].CorrID)
					assert.Equal(t, v.raw, records[i].Raw)
					assert.Equal(t, v.slug, records[i].Slug)
				}
				return
			}
			assert.True(t, len(records) == 0)
		})
	}
}

func BenchmarkFileRepo_Collect(b *testing.B) {
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

	repo, close, err := newFileStorage("bench_collect")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := close(); err != nil {
			b.Fatal(err)
		}
	}()

	for _, v := range shortenedURLs {
		if err := repo.Save(context.TODO(), v); err != nil {
			b.Fail()
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = repo.CollectByUser(context.TODO(), userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileRepo_Batch(t *testing.T) {
	type url struct {
		userID string
		corrID string
		raw    string
		slug   string
	}
	tests := []struct {
		name      string
		baseURL   string
		batchURLs []url
	}{
		{
			name:    "Batch",
			baseURL: "http://127.0.0.1/",
			batchURLs: []url{
				{
					userID: "2",
					corrID: "1",
					raw:    "http://demo2.com/",
					slug:   "slug21",
				},
				{
					userID: "3",
					corrID: "2",
					raw:    "http://demo3.com/",
					slug:   "slug32",
				},
				{
					userID: "4",
					corrID: "3",
					raw:    "http://demo4.com/",
					slug:   "slug14",
				},
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
					slug:   "slug1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, close, err := newFileStorage("test_batch")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			shotrenedURLs := make([]models.ShortenedURL, len(tt.batchURLs))
			for i, v := range tt.batchURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)
				shotrenedURLs[i] = shortenedURL
			}

			err = r.Batch(context.TODO(), shotrenedURLs)

			assert.NoError(t, err)
		})
	}
}

func TestFileRepo_Delete(t *testing.T) {
	r, close, err := newFileStorage("test_delete")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, close())
	}()

	err = r.Delete("0001", nil)
	require.Error(t, err)
}

func TestFileStorage_Stat(t *testing.T) {
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
			r, close, err := newFileStorage("test_getbyslug")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, close())
			}()

			for _, v := range tt.existedURLs {
				shortenedURL := models.NewShortenedURL(
					v.userID,
					v.corrID,
					v.raw,
					v.slug,
					tt.baseURL+v.slug,
				)

				err = r.Save(context.TODO(), shortenedURL)
				require.NoError(t, err)
			}

			got, err := r.Stat(context.TODO())
			require.NoError(t, err)
			assert.EqualValues(t, tt.want.URLs, got.URLs)
			assert.EqualValues(t, tt.want.Users, got.Users)
		})
	}
}

// newFileStorage creates a new fileStorage for test.
func newFileStorage(filename string) (*fileStorage, func() error, error) {
	// Prepare tmp filepath.
	start := time.Now()
	filename = fmt.Sprintf("%v\\%v%v", os.TempDir(), filename, strconv.FormatInt(start.Unix(), 10))

	repo, err := FileStorage(filename)
	if err != nil {
		return nil, nil, err
	}

	return repo, func() error {
		err1 := repo.Close()
		err2 := os.Remove(filename)
		if err1 == nil && err2 == nil {
			return nil
		}
		return fmt.Errorf("%v; %v", err1, err2)
	}, nil
}
