package models

import (
	"fmt"
)

// ShortenedURL represents the shortened URL.
type ShortenedURL struct {
	UserID    string
	CorrID    string
	Raw       string
	Slug      string
	Value     string
	IsDeleted bool
}

// NewShortenedURL returns a new ShortenedURL.
func NewShortenedURL(
	userID string,
	corrID string,
	raw string,
	slug string,
	value string,
) ShortenedURL {
	return ShortenedURL{
		UserID: userID,
		CorrID: corrID,
		Raw:    raw,
		Slug:   slug,
		Value:  value,
	}
}

// SetDeleted sets IsDeleted as true.
func (s *ShortenedURL) SetDeleted() {
	s.IsDeleted = true
}

// Empty checks on being empty.
func (s *ShortenedURL) Empty() bool {
	return len(s.UserID) == 0 &&
		len(s.CorrID) == 0 &&
		len(s.Raw) == 0 &&
		len(s.Value) == 0 &&
		len(s.Slug) == 0 &&
		!s.IsDeleted
}

// String represents ShortURL as a string.
func (s *ShortenedURL) String() string {
	return fmt.Sprintf("shortenedURL[userID: %s, corrID: %s, raw: %s, slug: %v, value: %s, deleted: %t]",
		s.UserID, s.CorrID, s.Raw, s.Slug, s.Value, s.IsDeleted)
}

// Equals compares ShortenedURLs.
func (s *ShortenedURL) Equals(s1 ShortenedURL) bool {
	return s.UserID == s1.UserID &&
		s.CorrID == s1.CorrID &&
		s.Raw == s1.Raw &&
		s.Value == s1.Value &&
		s.Slug == s1.Slug &&
		s.IsDeleted == s1.IsDeleted
}
