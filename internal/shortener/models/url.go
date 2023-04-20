package models

import (
	"fmt"
)

// URL is a representation of the URL that should be shortened.
type URL struct {
	UserID string
	CorrID string
	Raw    string
}

// NewURL returns a new URL.
func NewURL(userID string, corrID string, raw string) URL {
	return URL{
		UserID: userID,
		CorrID: corrID,
		Raw:    raw,
	}
}

// String returns URl as string.
func (u URL) String() string {
	return fmt.Sprintf("url[userID: %s, corrID: %s, value: %s]",
		u.UserID, u.CorrID, u.Raw)
}

// Equals compares URLs.
func (u URL) Equals(url URL) bool {
	return u.UserID == url.UserID &&
		u.CorrID == url.CorrID &&
		u.Raw == url.Raw
}

// Empty checks on being empty.
func (u URL) Empty() bool {
	return len(u.UserID) == 0 &&
		len(u.CorrID) == 0 &&
		len(u.Raw) == 0
}
