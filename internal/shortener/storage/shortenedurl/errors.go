package shortenedurl

import "errors"

// UniqueViolation while saving a new shortened URL.
var ErrUniqueViolation = errors.New("find shortened URL with the same raw URL")
