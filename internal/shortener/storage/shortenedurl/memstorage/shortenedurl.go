package memstorage

import "github.com/alukart32/shortener-url/internal/shortener/models"

// ShortURL represents shortened URL.
type shortenedURL struct {
	UserID    string
	CorrID    string
	Raw       string
	Value     string
	IsDeleted bool
}

// SetDeleted sets IsDeleted as true.
func (s *shortenedURL) SetDeleted() {
	s.IsDeleted = true
}

// newShortenedURL returns a new shortenedURL from models.ShortenedURL.
func newShortenedURL(s models.ShortenedURL) shortenedURL {
	return shortenedURL{
		UserID:    s.UserID,
		CorrID:    s.CorrID,
		Raw:       s.Raw,
		Value:     s.Value,
		IsDeleted: s.IsDeleted,
	}
}

// ToModel converts shortenedURL to models.ShortenedURL.
func (s *shortenedURL) ToModel(slug string) models.ShortenedURL {
	return models.ShortenedURL{
		UserID:    s.UserID,
		CorrID:    s.CorrID,
		Raw:       s.Raw,
		Slug:      slug,
		Value:     s.Value,
		IsDeleted: s.IsDeleted,
	}
}
