// Package memstorage defines a shortenedURL memory storage.
package memstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl"
)

// memStorage defines the shortenedURL storage in memory. It is based on a simple map.
type memStorage struct {
	data map[string]shortenedURL // slug: shortenedURL
	mtx  sync.RWMutex
}

// MemStorage returns a new empty memStorage.
func MemStorage() *memStorage {
	return &memStorage{
		data: make(map[string]shortenedURL, 32),
		mtx:  sync.RWMutex{},
	}
}

// GetBySlug finds the URL by slug.
func (ms *memStorage) GetBySlug(_ context.Context, slug string) (models.ShortenedURL, error) {
	if v, ok := ms.data[slug]; ok {
		return v.ToModel(slug), nil
	}

	return models.ShortenedURL{}, nil
}

// GetByURL returns a shortenedURL by original URL.
func (ms *memStorage) GetByURL(_ context.Context, url string) (models.ShortenedURL, error) {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	for slug, v := range ms.data {
		if v.Raw == url {
			return v.ToModel(slug), nil
		}
	}

	return models.ShortenedURL{}, nil
}

// Save saves the new URL in the repository.
func (ms *memStorage) Save(_ context.Context, data models.ShortenedURL) error {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	for _, v := range ms.data {
		if v.Raw == data.Raw && v.UserID == data.UserID {
			return shortenedurl.ErrUniqueViolation
		}
	}

	ms.data[data.Slug] = newShortenedURL(data)
	return nil
}

// CollectByUser collects user shortenedURLs.
func (ms *memStorage) CollectByUser(_ context.Context, userID string) ([]models.ShortenedURL, error) {
	if len(userID) == 0 {
		return nil, errors.New("userID is empty")
	}

	ms.mtx.RLock()
	records := []models.ShortenedURL{}
	defer ms.mtx.RUnlock()
	for slug, v := range ms.data {
		if v.UserID == userID {
			records = append(records, v.ToModel(slug))
		}
	}

	return records, nil
}

// Batch performs bulk shortenedURL insertion.
func (ms *memStorage) Batch(_ context.Context, shortURLs []models.ShortenedURL) error {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	for _, v := range shortURLs {
		ms.data[v.Slug] = newShortenedURL(v)
	}

	return nil
}

const workPoolSize = 10

// Delete marks urls as deleted.
func (ms *memStorage) Delete(userID string, slugs []string) error {
	// Set the appropriate pool size.
	wcount := workPoolSize
	if len(slugs) < wcount && len(slugs) >= 1 {
		wcount = len(slugs)
	}

	slugCh := make(chan string, wcount)
	go func() {
		defer close(slugCh)
		for _, s := range slugs {
			slugCh <- s
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < wcount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for slug := range slugCh {
				ms.mtx.Lock()
				if v, ok := ms.data[slug]; ok &&
					v.UserID == userID {
					v.SetDeleted()
					ms.data[slug] = v
				}
				ms.mtx.Unlock()
			}
		}()
	}
	wg.Wait()

	return nil
}

// Stat collects statistics about shortened URLs.
func (ms *memStorage) Stat(_ context.Context) (models.Stat, error) {
	var stat models.Stat

	done := make(chan struct{}, 1)
	go func() {
		ms.mtx.RLock()
		stat.URLs = len(ms.data)
		users := make(map[string]struct{})
		for _, v := range ms.data {
			users[v.UserID] = struct{}{}
		}
		stat.Users = len(users)
		ms.mtx.RUnlock()

		done <- struct{}{}
	}()

	<-done
	return stat, nil
}
