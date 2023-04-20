// Package filestorage defines defines a shortenedURL file storage based on msgp encoding.
package filestorage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/caarlos0/env/v6"
)

// config represents the file storage configuration.
type config struct {
	URI string
}

// fileStorage defines the shortenedURL file storage.
type fileStorage struct {
	w   writer
	r   reader
	mtx sync.Mutex
}

// New returns a new fileStorage for shortenedURLs.
func FileStorage(path string) (*fileStorage, error) {
	if len(path) == 0 {
		var cfg config
		opts := env.Options{RequiredIfNoDef: true}
		err := env.Parse(&cfg, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to set fileStorage")
		}
		path = cfg.URI
	}

	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	fr, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &fileStorage{
		w: newMsgpWriter(fw),
		r: newMsgpReader(fr),
	}, nil
}

// GetBySlug finds the shortURL by slug.
func (fs *fileStorage) GetBySlug(_ context.Context, slug string) (models.ShortenedURL, error) {
	fs.mtx.Lock()
	var records []ShortenedURL
	records, err := fs.r.List()
	fs.mtx.Unlock()

	if err != nil {
		return models.ShortenedURL{}, err
	}

	for _, r := range records {
		if r.Slug == slug {
			return r.ToModel(), nil
		}
	}

	return models.ShortenedURL{}, nil
}

// GetByURL finds the shortURL by original URL.
func (fs *fileStorage) GetByURL(_ context.Context, url string) (models.ShortenedURL, error) {
	fs.mtx.Lock()
	var records []ShortenedURL
	records, err := fs.r.List()
	fs.mtx.Unlock()

	if err != nil {
		return models.ShortenedURL{}, err
	}

	for _, r := range records {
		if r.Raw == url {
			return r.ToModel(), nil
		}
	}
	return models.ShortenedURL{}, nil
}

// Collect finds all user shortenedURLs.
func (fs *fileStorage) CollectByUser(_ context.Context, userID string) ([]models.ShortenedURL, error) {
	if len(userID) == 0 {
		return nil, errors.New("userID is empty")
	}

	fs.mtx.Lock()
	records, err := fs.r.List()
	fs.mtx.Unlock()

	if err != nil {
		return nil, err
	}

	var idxs []int
	for i, r := range records {
		if r.UserID == userID {
			idxs = append(idxs, i)
		}
	}

	results := make(shortenedURLs, len(idxs))
	for i, idx := range idxs {
		results[i] = records[idx]
	}
	return results.ToModel(), nil
}

// Save saves a new shortURL.
func (fs *fileStorage) Save(_ context.Context, data models.ShortenedURL) error {
	fs.mtx.Lock()
	err := fs.w.Write(newShortenedURL(data))
	fs.mtx.Unlock()
	return err
}

// Batch saves a list of shortenedURLs.
func (fs *fileStorage) Batch(_ context.Context, records []models.ShortenedURL) error {
	fs.mtx.Lock()
	for _, r := range records {
		err := fs.w.Write(newShortenedURL(r))
		if err != nil {
			// BUG: No Rollback. Added entries should be removed.
			return nil
		}
	}
	fs.mtx.Unlock()
	return nil
}

// Delete marks urls as deleted.
func (fs *fileStorage) Delete(userID string, slugs []string) error {
	// FYI: Do nothing for this version.
	// Should use the LSM-tree for updating entries in the log segment.
	return fmt.Errorf("not implemented")
}

// Stat collects statistics about shortened URLs.
func (fs *fileStorage) Stat(_ context.Context) (models.Stat, error) {
	var (
		stat models.Stat
		err  error
	)

	done := make(chan struct{}, 1)
	go func() {
		var records []ShortenedURL

		fs.mtx.Lock()
		records, err = fs.r.List()
		fs.mtx.Unlock()

		if err != nil {
			return
		}

		users := make(map[string]struct{})
		for _, v := range records {
			users[v.UserID] = struct{}{}
		}
		stat.Users = len(users)
		stat.URLs = len(records)

		done <- struct{}{}
	}()
	<-done
	return stat, err
}

// Close closes the writer and reader.
func (fs *fileStorage) Close() error {
	err1 := fs.w.Close()
	err2 := fs.r.Close()

	if err1 != nil || err2 != nil {
		return fmt.Errorf("close files - %v; %v", err1, err2)
	}
	return nil
}
