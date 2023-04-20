package filestorage

import (
	"errors"
	"io"
	"os"

	"github.com/tinylib/msgp/msgp"
)

// reader defines the file reader.
type reader interface {
	List() ([]ShortenedURL, error)
	io.Closer
}

var _ reader = (*msgpReader)(nil)

// msgpReader is a msgp file reader.
type msgpReader struct {
	file   *os.File
	reader *msgp.Reader
}

// newMsgpReader returns a new msgpReader.
func newMsgpReader(f *os.File) reader {
	return &msgpReader{
		file:   f,
		reader: msgp.NewReader(f),
	}
}

// List returns all entries in the file. A successful call returns err == nil.
func (r *msgpReader) List() (records []ShortenedURL, err error) {
	defer func() {
		err = r.rewind()
	}()

	var record ShortenedURL
	err = record.DecodeMsg(r.reader)
	for err == nil {
		records = append(records, record)
		err = record.DecodeMsg(r.reader)
	}
	if err != nil && errors.Unwrap(err) != io.EOF {
		return nil, err
	}

	return records, nil
}

// Close closes msgpReader.
func (r *msgpReader) Close() error {
	return r.file.Close()
}

// rewind sets the search position at the beginning of the file and updates reader.
func (r *msgpReader) rewind() error {
	_, err := r.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	r.reader.Reset(r.file)
	return err
}
