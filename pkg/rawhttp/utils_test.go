package rawhttp

import (
	"errors"
)

var ErrForTestWrite = errors.New("ErrorWriter.Write: force error for tests")

type ErrorWriter struct {
	count uint64
	when  uint64
}

func (w *ErrorWriter) Write(p []byte) (int, error) {
	w.count++

	if w.when == 0 || w.when == w.count {
		return 0, ErrForTestWrite
	}

	return len(p), nil
}
