// Package shorturl provides URL shortener functionality.
package shorturl

import (
	"context"
	"errors"
	"fmt"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl"
	"github.com/caarlos0/env/v6"
)

// config represents the shortener configuration.
type config struct {
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

// urlSaver defines a urlSaver for the shortened URL.
type urlSaver interface {
	Save(context.Context, models.ShortenedURL) error
	Batch(context.Context, []models.ShortenedURL) error
}

// shortener is a representation of the url shortener.
type shortener struct {
	saver   urlSaver
	baseURL string
}

// NewShortener returns a new shortener.
func Shortener(baseURL string, saver urlSaver) (*shortener, error) {
	if saver == nil {
		return nil, fmt.Errorf("url saver is nil")
	}

	if len(baseURL) == 0 {
		var cfg config
		opts := env.Options{RequiredIfNoDef: true}
		err := env.Parse(&cfg, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to set config")
		}
		baseURL = cfg.BaseURL
	}

	return &shortener{
		baseURL: baseURL,
		saver:   saver,
	}, nil
}

// URL shortening errors.
var (
	ErrInternal        = errors.New("internal error")
	ErrInvalidCreation = errors.New("invalid creation")
	ErrUniqueViolation = errors.New("unique violation")
)

// Short creates and saves a new shortened URL.
func (s *shortener) Short(ctx context.Context, url models.URL) (string, error) {
	shortenedURL, err := shortenURL(
		url.UserID,
		url.CorrID,
		url.Raw,
		s.baseURL,
	)
	if err != nil {
		return "", ErrInvalidCreation
	}

	if err = s.saver.Save(ctx, shortenedURL.ToModel()); err != nil {
		if errors.Is(err, shortenedurl.ErrUniqueViolation) {
			err = ErrUniqueViolation
		}
		return "", err
	}

	return shortenedURL.Value, nil
}

// Batch was for empty URLs list.
var ErrEmptyBatch = errors.New("empty batch")

// Batch creates and saves a new shortened URLs.
func (s *shortener) Batch(ctx context.Context, urls []models.URL) ([]models.ShortenedURL, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyBatch
	}

	shortenedURLs := make(shortenedURLs, len(urls))
	for i, v := range urls {
		s, err := shortenURL(
			v.UserID,
			v.CorrID,
			v.Raw,
			s.baseURL,
		)
		if err != nil {
			return nil, ErrInvalidCreation
		}
		shortenedURLs[i] = s
	}

	urlsToBatch := shortenedURLs.ToModel()
	return urlsToBatch, s.saver.Batch(ctx, urlsToBatch)
}
