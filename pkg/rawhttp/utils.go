package rawhttp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
)

// ReadUntilDelimiter will read data from the given reader until the delimiter is found.
// The maximal length of the returned data is determinated by the reader buffer len.
func ReadUntilDelimiter(b *bufio.Reader, delim []byte) ([]byte, error) {
	// Ensure that the delimiter is not empty
	_ = delim[0]
	delimLen := len(delim)

	n := delimLen
	for {
		raw, err := b.Peek(n)

		if i := bytes.Index(raw, delim); i >= 0 {
			toDiscard := i + delimLen
			if discarded, discardErr := b.Discard(toDiscard); discardErr != nil {
				panic("ReadUntilDelimiter: cannot discard current line: " + discardErr.Error())
			} else if toDiscard != discarded {
				panic(fmt.Sprintf("ReadUntilDelimiter: cannot discard enough bytes... %d != %d", toDiscard, discarded))
			}
			return slices.Clone(raw[:i]), nil
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return nil, fmt.Errorf("ReadUntilDelimiter: peek error: %w", err)
		}

		n = b.Buffered() + 1
	}
}
