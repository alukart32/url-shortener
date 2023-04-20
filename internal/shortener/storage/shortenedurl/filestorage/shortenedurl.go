package filestorage

import "github.com/alukart32/shortener-url/internal/shortener/models"

//go:generate msgp

// ShortenedURL is a shortened URL in the file storage.
type ShortenedURL struct {
	Slug      string `msg:"slug"`
	UserID    string `msg:"userID"`
	CorrID    string `msg:"corrID"`
	Value     string `msg:"value"`
	Raw       string `msg:"Raw"`
	IsDeleted bool   `msg:"is_deleted"`
}

// newShortenedURL returns a new ShortenedURL from model.
func newShortenedURL(s models.ShortenedURL) ShortenedURL {
	return ShortenedURL{
		UserID:    s.UserID,
		CorrID:    s.CorrID,
		Raw:       s.Raw,
		Slug:      s.Slug,
		Value:     s.Value,
		IsDeleted: s.IsDeleted,
	}
}

// ToModel converts ShortenedURL to model.ShortenedURL.
func (s *ShortenedURL) ToModel() models.ShortenedURL {
	return models.ShortenedURL{
		UserID:    s.UserID,
		CorrID:    s.CorrID,
		Raw:       s.Raw,
		Slug:      s.Slug,
		Value:     s.Value,
		IsDeleted: s.IsDeleted,
	}
}

// shortenedURLs is the set of ShortenedURL.
type shortenedURLs []ShortenedURL

// ToModel converts ShortenedURLs to model.ShortenedURLs.
func (s shortenedURLs) ToModel() []models.ShortenedURL {
	tmp := make([]models.ShortenedURL, len(s))
	for i, v := range s {
		tmp[i] = v.ToModel()
	}

	return tmp
}
