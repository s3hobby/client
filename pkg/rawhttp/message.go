package rawhttp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"slices"
	"strings"
)

type Message struct {
	StartLine StartLine
	Header    Header
}

func (req *Message) Marshal(w io.Writer) error {
	if err := req.StartLine.Marshal(w); err != nil {
		return fmt.Errorf("Message.Marshal: cannot marshal start line: %w", err)
	}

	if err := req.Header.Marshal(w); err != nil {
		return fmt.Errorf("Message.Marshal: cannot marshal header: %w", err)
	}

	return nil
}

func (req *Message) Unmarshal(b *bufio.Reader) error {
	if err := req.StartLine.Unmarshal(b); err != nil {
		return err
	}

	if err := req.Header.Unmarshal(b); err != nil {
		return err
	}

	return nil
}

type StartLine struct {
	First  string
	Second string
	Third  string
}

func (sl *StartLine) Unmarshal(b *bufio.Reader) error {
	if first, err := ReadUntilDelimiter(b, []byte{' '}); err != nil {
		return fmt.Errorf("StartLine.Unmarshal: cannot read first element: %w", err)
	} else if len(first) == 0 {
		return errors.New("StartLine.Unmarshal: first element is empty")
	} else {
		sl.First = string(first)
	}

	if second, err := ReadUntilDelimiter(b, []byte{' '}); err != nil {
		return fmt.Errorf("StartLine.Unmarshal: cannot read second element: %w", err)
	} else if len(second) == 0 {
		return errors.New("StartLine.Unmarshal: second element is empty")
	} else {
		sl.Second = string(second)
	}

	if third, err := ReadUntilDelimiter(b, []byte{'\r', '\n'}); err != nil {
		return fmt.Errorf("StartLine.Unmarshal: cannot read third element: %w", err)
	} else if len(third) == 0 {
		return errors.New("StartLine.Unmarshal: third element is empty")
	} else {
		sl.Third = string(third)
	}

	return nil
}

func (sl *StartLine) Marshal(w io.Writer) error {
	line := []byte(sl.String())

	_, err := w.Write(line)
	if err != nil {
		return fmt.Errorf("StartLine.Marshal: cannot write: %w", err)
	}

	return nil
}

func (sl *StartLine) String() string {
	return fmt.Sprintf(
		"%s %s %s\r\n",
		sl.First,
		sl.Second,
		sl.Third,
	)
}

type Header struct {
	data map[string]string
}

func (h *Header) Has(key string) bool {
	_, ok := h.data[strings.ToLower(key)]
	return ok
}

func (h *Header) Set(key, value string) {
	if h.data == nil {
		h.data = make(map[string]string)
	}

	h.data[strings.ToLower(key)] = value
}

func (h *Header) Get(key string) (value string, exists bool) {
	value, exists = h.data[strings.ToLower(key)]
	return
}

func (h *Header) Del(key string) {
	delete(h.data, strings.ToLower(key))
}

func (h *Header) Keys() iter.Seq[string] {
	return maps.Keys(h.data)
}

func (h *Header) All() iter.Seq2[string, string] {
	return maps.All(h.data)
}

func (h *Header) Marshal(w io.Writer) error {
	keys := slices.Sorted(maps.Keys(h.data))
	for _, key := range keys {
		line := []byte(fmt.Sprintf("%s: %s\r\n", key, h.data[key]))
		if _, err := w.Write(line); err != nil {
			return fmt.Errorf("Header.Marshal: cannot write line: %w", err)
		}
	}

	if _, err := w.Write([]byte("\r\n")); err != nil {
		return fmt.Errorf("Header.Marshal: cannot write last line: %w", err)
	}

	return nil
}

func (h *Header) Unmarshal(r *bufio.Reader) error {
	const eolValue = "\r\n"
	const eolLen = len(eolValue)

	for {
		peek, err := r.Peek(eolLen)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return fmt.Errorf("Header.Unmarshal: cannot peek the first %d bytes of the line: %w", eolLen, err)
		}
		if bytes.Equal(peek, []byte(eolValue)) {
			if discarded, discardErr := r.Discard(eolLen); discardErr != nil {
				panic(fmt.Sprintf("Header.Unmarshal: cannot the first %d bytes: %v", eolLen, discardErr))
			} else if eolLen != discarded {
				panic(fmt.Sprintf("Header.Unmarshal: cannot discard enough bytes... %d != %d", eolLen, discarded))
			}
			return nil
		}

		name, err := ReadUntilDelimiter(r, []byte{':'})
		if err != nil {
			return fmt.Errorf("Header.Unmarshal: cannot read header name: %w", err)
		} else if len(name) == 0 {
			return errors.New("Header.Unmarshal: header name is empty")
		}

		value, err := ReadUntilDelimiter(r, []byte(eolValue))
		if err != nil {
			return fmt.Errorf("Header.Unmarshal: cannot read header value: %w", err)
		}
		value = bytes.TrimSpace(value)
		if len(value) == 0 {
			return errors.New("Header.Unmarshal: header value is empty")
		}

		if h.Has(string(name)) {
			return fmt.Errorf("Headers.Unmarshal: header %q already set", name)
		}

		h.Set(string(name), string(value))
	}
}
