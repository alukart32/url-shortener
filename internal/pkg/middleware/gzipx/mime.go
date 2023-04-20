package gzipx

import "strings"

var mimes = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

// isCompressibleMIME verifies the mime for compression.
func isCompressibleMIME(mime string) bool {
	if len(mime) == 0 {
		return false
	}

	for _, v := range mimes {
		if strings.Contains(v, mime) {
			return true
		}
	}
	return false
}
