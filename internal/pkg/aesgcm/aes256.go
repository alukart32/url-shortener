// Package aesgcm provides AES encryption with Galois/Counter Mode (GCM).
package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
)

// ErrInvalidValue represents invalid a value encoding/decoding error.
var ErrInvalidValue = errors.New("invalid value")

// Key256 represents a 256-bit aes key.
type Key256 [32]byte

// HashKey256 gets the sha256 hash for the key.
func HashKey256(key string) [32]byte {
	return sha256.Sum256([]byte(key))
}

// Seal seals data in the format "{nonce}{encrypted plaintext}".
func Seal(data string, key Key256) ([]byte, error) {
	gcm, err := gcmBlock(key[:])
	if err != nil {
		return nil, err
	}

	nonce, err := randBytes(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	// By passing the nonce as the first parameter, the encrypted data
	// will be appended to the nonce — meaning that the returned encryptedValue
	// will be in the format "{nonce}{encrypted plaintext}".
	return gcm.Seal(nonce, nonce, []byte(data), nil), nil
}

// Open opens a sealed message by using 256-bits key.
func Open(data string, key Key256) ([]byte, error) {
	gcm, err := gcmBlock(key[:])
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidValue
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// Decrypt and authenticate the ciphertext.
	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return nil, ErrInvalidValue
	}

	return plaintext, nil
}

// gcmBlock creates a new GCM block (AEAD cipher mode).
func gcmBlock(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}

// randBytes generates random bytes. Size defines the length of the slice.
// A successful call returns err == nil.
func randBytes(size int) ([]byte, error) {
	if size == 0 {
		return nil, fmt.Errorf("zero size")
	}

	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
