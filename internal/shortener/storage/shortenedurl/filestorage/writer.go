package filestorage

import (
	"io"
	"os"

	"github.com/tinylib/msgp/msgp"
)

// writer defines the file writer.
type writer interface {
	Write(ShortenedURL) error
	io.Closer
}

var _ writer = (*msgpWriter)(nil)

// msgpackWriter is a msgp file writer.
type msgpWriter struct {
	file   *os.File
	writer *msgp.Writer
}

// newMsgpWriter returns a new msgpWriter.
func newMsgpWriter(f *os.File) writer {
	return &msgpWriter{
		file:   f,
		writer: msgp.NewWriter(f),
	}
}

// Write writes a new msgp entry down. A successful call returns err == nil.
func (w *msgpWriter) Write(data ShortenedURL) error {
	b, err := data.MarshalMsg(nil)
	if err != nil {
		return err
	}

	_, err = w.file.Write(b)
	return err
}

// Close closes msgpWriter.
func (w *msgpWriter) Close() error {
	return w.file.Close()
}
