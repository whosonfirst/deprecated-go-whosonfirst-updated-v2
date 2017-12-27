package writer

import (
       "io"
)

type NullWriter struct {
	Writer
}

func NewNullWriter() (Writer, error) {

	w := NullWriter{}
	return &w, nil
}

func Write(fh io.ReadCloser) error {
	// maybe drain fh here?
	return nil
}
