// Package jwttoken provides JWT manager.
package jwttoken

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/golang-jwt/jwt/v4"
)

// Manager represents JWT token manager.
type manager struct {
	key    string
	expDur time.Duration
}

// Manager returns a new JWT token manager.
func Manager(cfg Config) (*manager, error) {
	if cfg.Empty() {
		opts := env.Options{RequiredIfNoDef: true}
		if err := env.Parse(&cfg, opts); err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
	}

	return &manager{
		key:    cfg.Key,
		expDur: time.Duration(cfg.ExpTime),
	}, nil
}

// NewToken generates a new JWT token based on userID. It returns the token as a string.
func (m *manager) NewToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &authClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expDur)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(m.key))
}

// VerifyToken validates and returns the parsed userID.
func (m *manager) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.key), nil
	})
	if err != nil {
		return "", fmt.Errorf("JWT parse: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["user_id"].(string), nil
	} else {
		return "", err
	}
}

// authClaims defines JWT authClaims with with userID.
type authClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// Valid validates claims.
func (c authClaims) Valid() error {
	err := c.RegisteredClaims.Valid()
	if err != nil {
		return err
	}

	if len(c.UserID) == 0 {
		return fmt.Errorf("empty user_id")
	}

	return nil
}

// Config represents the JWT manager configuration.
type Config struct {
	Key     string        `env:"JWT_SIGN_KEY" envDefault:"secret"`
	ExpTime time.Duration `env:"JWT_TOKEN_EXPR" envDefault:"1h"`
}

// Empty checks on being empty.
func (c Config) Empty() bool {
	return len(c.Key) == 0 && c.ExpTime == 0
}
