package httpauth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alukart32/shortener-url/internal/pkg/aesgcm"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookieWrapper_ReqWithCookie(t *testing.T) {
	cookieAuth := cookieAuthProvider{
		key:      "topsecret",
		name:     "user_id",
		maxAge:   0,
		secure:   false,
		httpOnly: true,
	}

	// set router.
	userID := "1"
	router := gin.New()
	router.POST("/", cookieAuth.Handle(func(c *gin.Context, userID string) {
		c.String(200, userID)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	// encode cookie value.
	cookie := http.Cookie{
		Name:  cookieAuth.name,
		Value: userID,
	}
	require.NoError(t, testEnc(&cookie, cookieAuth.key))
	req.AddCookie(&cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	result := w.Result()
	// there should be no cookie in the response.
	assert.True(t, len(result.Cookies()) == 0)
	assert.Equal(t, userID, w.Body.String())
	require.NoError(t, result.Body.Close())
}

func TestCookieWrapper_ReqWithoutCookie(t *testing.T) {
	cookieAuth := cookieAuthProvider{
		key:      "topsecret",
		name:     "user_id",
		maxAge:   0,
		secure:   false,
		httpOnly: true,
	}

	// set router.
	router := gin.New()
	router.POST("/", cookieAuth.Handle(func(c *gin.Context, userID string) {
		c.String(200, userID)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	result := w.Result()

	// decrypt cookie value.
	cookie := result.Cookies()[0]
	value, err := testDecode(cookie, cookieAuth.key)
	require.NoError(t, err)
	assert.Equal(t, w.Body.String(), value)
	require.NoError(t, result.Body.Close())
}

func testEnc(cookie *http.Cookie, key string) error {
	plaintext := cookie.Name + ":" + cookie.Value
	// EncryptedValue is in the format "{nonce}{encrypted plaintext}"
	encryptedValue, err := aesgcm.Seal(plaintext, aesgcm.HashKey256(key))
	if err != nil {
		return err
	}
	cookie.Value = string(encryptedValue)

	cookie.Value = base64.RawURLEncoding.EncodeToString([]byte(cookie.Value))

	if len(cookie.String()) > 4096 {
		return ErrLongCookie
	}
	return nil
}

func testDecode(cookie *http.Cookie, key string) (string, error) {
	// Decode the base64-encoded cookie value.
	decoded, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}

	encryptedValue := string(decoded)

	plaintext, err := aesgcm.Open(encryptedValue, aesgcm.HashKey256(key))
	if err != nil {
		return "", err
	}

	// The plaintext value is in the format "{name}:{value}".
	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		return "", ErrInvalidCookie
	}

	if expectedName != cookie.Name {
		return "", ErrInvalidCookie
	}

	// Return the plaintext cookie value.
	return string(value), nil
}
