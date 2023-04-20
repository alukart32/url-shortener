package gzipx

import "testing"

func TestMIME(t *testing.T) {
	tests := []struct {
		name     string
		mime     string
		notFound bool
	}{
		{
			name: "Valid application/json mime",
			mime: "application/json",
		},
		{
			name: "Valid text/plain mime",
			mime: "text/plain",
		},
		{
			name:     "Invalid text/oct mime",
			mime:     "text/oct",
			notFound: true,
		},
		{
			name:     "Empty mime",
			mime:     "",
			notFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := isCompressibleMIME(tt.mime); !ok && !tt.notFound {
				t.Errorf("should find %v", tt.mime)
			}
		})
	}
}
