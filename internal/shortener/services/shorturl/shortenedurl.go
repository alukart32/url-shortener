package shorturl

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/alukart32/shortener-url/internal/shortener/models"
)

// shortenedURL represents the shortened URL.
type shortenedURL struct {
	UserID    string
	CorrID    string
	Raw       string
	Slug      string
	Value     string
	IsDeleted bool
}

// shortenURL returns a new shortenedURL.
func shortenURL(
	userID string,
	corrID string,
	rawURI string,
	baseURL string,
) (shortenedURL, error) {
	if len(userID) == 0 {
		return shortenedURL{}, fmt.Errorf("empty userID")
	}
	if len(rawURI) == 0 {
		return shortenedURL{}, fmt.Errorf("empty URI")
	}

	uri, err := url.ParseRequestURI(rawURI)
	if err != nil {
		return shortenedURL{}, fmt.Errorf("failed to parse URI")
	}
	if addr := net.ParseIP(uri.Host); addr == nil {
		if !strings.Contains(uri.Host, ".") {
			return shortenedURL{}, fmt.Errorf("failed to parse URI")
		}
	}
	if len(baseURL) == 0 {
		return shortenedURL{}, fmt.Errorf("empty baseURL")
	}

	slug, err := base62(7)
	if err != nil {
		return shortenedURL{}, fmt.Errorf("failed to create the slug: %v", err)
	}

	return shortenedURL{
		UserID: userID,
		CorrID: corrID,
		Raw:    rawURI,
		Slug:   slug,
		Value:  baseURL + "/" + slug,
	}, nil
}

// SetDeleted sets IsDeleted as true.
func (s *shortenedURL) SetDeleted() {
	s.IsDeleted = true
}

// Empty checks on being empty.
func (s *shortenedURL) Empty() bool {
	return len(s.UserID) == 0 &&
		len(s.CorrID) == 0 &&
		len(s.Raw) == 0 &&
		len(s.Value) == 0 &&
		len(s.Slug) == 0 &&
		!s.IsDeleted
}

// String represents ShortURL as a string.
func (s *shortenedURL) String() string {
	return fmt.Sprintf("shortenedURL[userID: %s, corrID: %s, raw: %s, slug: %v, value: %s, deleted: %t]",
		s.UserID, s.CorrID, s.Raw, s.Slug, s.Value, s.IsDeleted)
}

// Equals compares ShortURL.
func (s *shortenedURL) Equals(s1 shortenedURL) bool {
	return s.UserID == s1.UserID &&
		s.CorrID == s1.CorrID &&
		s.Raw == s1.Raw &&
		s.Value == s1.Value &&
		s.Slug == s1.Slug &&
		s.IsDeleted == s1.IsDeleted
}

// ToModel converts shortenedURL to models.ShortenedURL.
func (s *shortenedURL) ToModel() models.ShortenedURL {
	return models.ShortenedURL{
		UserID:    s.UserID,
		CorrID:    s.CorrID,
		Raw:       s.Raw,
		Slug:      s.Slug,
		Value:     s.Value,
		IsDeleted: s.IsDeleted,
	}
}

// shortenedURLs represents a slice of shortenedURL.
type shortenedURLs []shortenedURL

// ToModel converts shortenedURLs to models.ShortenedURLs.
func (s shortenedURLs) ToModel() []models.ShortenedURL {
	tmp := make([]models.ShortenedURL, len(s))
	for i, v := range s {
		tmp[i] = v.ToModel()
	}

	return tmp
}

const (
	base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base62Len      = uint64(len(base62Alphabet))
)

// base62 returns base62 encoded strings.
func base62(size int) (string, error) {
	if size == 0 {
		return "", fmt.Errorf("zero size")
	}

	var sb strings.Builder
	sb.Grow(size)

	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.New("failed to generate random uint64")
	}

	base := binary.LittleEndian.Uint64(randomBytes[:])
	for i := 0; i < size; i++ {
		base = base / base62Len
		sb.WriteByte(base62Alphabet[(base % base62Len)])
	}

	return sb.String(), nil
}
