// Package services defines application services for presentation/controller layer.
package services

import (
	"context"

	"github.com/alukart32/shortener-url/internal/shortener/models"
)

// Shortener defines the url Shortener.
type Shortener interface {
	Short(context.Context, models.URL) (string, error)
	Batch(context.Context, []models.URL) ([]models.ShortenedURL, error)
}

// Provider defines the shortened URL Provider.
type Provider interface {
	GetByURL(context.Context, string) (models.ShortenedURL, error)
	GetBySlug(context.Context, string) (models.ShortenedURL, error)
	CollectByUser(context.Context, string) ([]models.ShortenedURL, error)
}

// Deleter defines the shortURLs Deleter by slug.
type Deleter interface {
	Delete(userID string, slugs []string) error
}

// StatProvider defines the shortened URLs statistics provider.
type StatProvider interface {
	Stat(context.Context) (models.Stat, error)
}

// Pinger defines the network pinger.
type Pinger interface {
	Ping() error
}

// Services is a representation of all application services.
type Services struct {
	Shortener      Shortener
	Provider       Provider
	Deleter        Deleter
	Statistic      StatProvider
	PostgresPinger Pinger
}

// NewServices combines a new application Servies.
func NewServices(
	shortener Shortener,
	provider Provider,
	deleter Deleter,
	statistic StatProvider,
	pgxPinger Pinger,
) *Services {
	return &Services{
		Shortener:      shortener,
		Provider:       provider,
		Deleter:        deleter,
		Statistic:      statistic,
		PostgresPinger: pgxPinger,
	}
}
