package rawhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
