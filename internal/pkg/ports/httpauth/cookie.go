// Package httpauth provides an authentication provider for gin.HandlerFunc.
package httpauth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/alukart32/shortener-url/internal/pkg/aesgcm"
	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// cookieAuthProvider represents the cookie auth provider.
type cookieAuthProvider struct {
	key      string
	name     string
	maxAge   int
	secure   bool
	httpOnly bool
}

// CookieAuthProvider returns a new cookieAuthProvider. Error will return only if the cookie configuration cannot be loaded.
func CookieAuthProvider() (*cookieAuthProvider, error) {
	cfg, err := newCookieConfig()
	if err != nil {
		return nil, err
	}
	return &cookieAuthProvider{
		key:      cfg.Key,
		name:     cfg.Name,
		maxAge:   cfg.MaxAge,
		secure:   cfg.Secure,
		httpOnly: cfg.HTTPOnly,
	}, nil
}

// cookieConfig represents the cookie auth wrapper configuration.
type cookieConfig struct {
	Key      string `env:"AUTH_COOKIE_KEY_PATH" envDefault:"bla-bla"`
	Name     string `env:"AUTH_COOKIE_NAME" envDefault:"user_id"`
	MaxAge   int    `env:"AUTH_COOKIE_MAX_AGE" envDefault:"0"`
	Secure   bool   `env:"AUTH_COOKIE_SECURE" envDefault:"false"`
	HTTPOnly bool   `env:"AUTH_COOKIE_HTTP" envDefault:"true"`
}

// newCookieConfig returns a new config.
func newCookieConfig() (*cookieConfig, error) {
	opts := env.Options{RequiredIfNoDef: true}

	var cfg cookieConfig
	err := env.Parse(&cfg, opts)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Handle adds cookie auth functionality to .gin.HandlerFunc.
func (cw *cookieAuthProvider) Handle(next func(c *gin.Context, userID string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := cw.readEncrypted(c, cw.name, cw.key)
		if err == nil {
			next(c, userID)
			return
		}

		// Generate a new userID.
		id, err := uuid.NewUUID()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Add a cookie for the request.
		c.Request.Header.Add(cw.name, id.String())
		// Add a cookie for the response.
		cookie := &http.Cookie{
			Name:     cw.name,
			Value:    id.String(),
			MaxAge:   cw.maxAge,
			Secure:   cw.secure,
			HttpOnly: cw.httpOnly,
		}
		cw.writeEncrypted(c, cookie, cw.key)

		next(c, id.String())
	}
}

// Cookie encode/decode errors.
var (
	ErrLongCookie    = errors.New("too long cookie")
	ErrInvalidCookie = errors.New("invalid cookie")
)

// write writes a base64 encoded value to a cookie.
func (cw *cookieAuthProvider) write(c *gin.Context, cookie *http.Cookie) error {
	// Encode the cookie value using base64.
	cookie.Value = base64.RawURLEncoding.EncodeToString([]byte(cookie.Value))

	// Check the total length of the cookie contents.
	// Return the error if it's more than 4096 bytes.
	if len(cookie.String()) > 4096 {
		return ErrLongCookie
	}

	// Write the cookie.
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge,
		cookie.Path, cookie.Domain, false, true)
	return nil
}

// writeEncrypted writes an aes encrypted value to a cookie.
func (cw *cookieAuthProvider) writeEncrypted(c *gin.Context, cookie *http.Cookie, key string) error {
	// EncryptedValue is in the format "{nonce}{encrypted plaintext}"
	encryptedValue, err := aesgcm.Seal(cookie.Name+":"+cookie.Value, aesgcm.HashKey256(key))
	if err != nil {
		return err
	}
	cookie.Value = string(encryptedValue)
	return cw.write(c, cookie)
}

// read reads the base64 encoded value from the cookie.
func (cw *cookieAuthProvider) read(c *gin.Context, name string) (string, error) {
	cookie, err := c.Cookie(name)
	if err != nil {
		return "", err
	}

	// Decode the base64-encoded cookie value.
	value, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// readEncrypted reads an aes encrypted value from a cookie.
func (cw *cookieAuthProvider) readEncrypted(c *gin.Context, name string, key string) (string, error) {
	// Read in the signed value from the cookie in the format
	// "{nonce}{encrypted plaintext}".
	encryptedValue, err := cw.read(c, name)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(encryptedValue, aesgcm.HashKey256(key))
	if err != nil {
		return "", err
	}

	// The plaintext value is in the format "{name}:{value}".
	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok || expectedName != name {
		return "", ErrInvalidCookie
	}

	// Return the plaintext cookie value.
	return value, nil
}
