package rawhttp

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	newHeader := func(t *testing.T, raw map[string]string) *Header {
		h := &Header{}

		for k, v := range raw {
			if h.Has(k) {
				require.Failf(t, "Header %q already present", k)
			}

			h.Set(k, v)
		}

		return h
	}

	t.Run("marshal", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			testCases := []struct {
				name     string
				input    map[string]string
				expected string
			}{
				{
					name:     "empty",
					input:    nil,
					expected: "\r\n",
				},
				{
					name:     "one value",
					input:    map[string]string{"my-key": "my-value"},
					expected: "my-key: my-value\r\n\r\n",
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					var buf bytes.Buffer
					h := newHeader(t, tc.input)
					err := h.Marshal(&buf)
					require.NoError(t, err)
					require.Equal(t, tc.expected, buf.String())
				})
			}
		})

		t.Run("error", func(t *testing.T) {
			t.Run("line write error", func(t *testing.T) {
				data := map[string]string{
					"the-key": "the-value",
				}
				h := newHeader(t, data)
				err := h.Marshal(&ErrorWriter{})
				require.ErrorIs(t, err, ErrForTestWrite)
			})

			t.Run("ending write error", func(t *testing.T) {
				h := &Header{}
				err := h.Marshal(&ErrorWriter{})
				require.ErrorIs(t, err, ErrForTestWrite)
			})
		})
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			testCases := []struct {
				name     string
				input    string
				expected map[string]string
				trailer  string
			}{
				{
					name:     "empty",
					input:    "\r\n",
					expected: map[string]string{},
				},
				{
					name:     "one value",
					input:    "the-key: the-value\r\n\r\n",
					expected: map[string]string{"the-key": "the-value"},
				},
				{
					name:  "two values",
					input: "01234: 56789\r\nabcde: fghij\r\n\r\n",
					expected: map[string]string{
						"01234": "56789",
						"abcde": "fghij",
					},
				},
				{
					name:     "trailer",
					input:    "the-key: the-value\r\n\r\nthe\r\ntrailer",
					expected: map[string]string{"the-key": "the-value"},
					trailer:  "the\r\ntrailer",
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					r := bufio.NewReader(bytes.NewReader([]byte(tc.input)))

					header := &Header{}
					err := header.Unmarshal(r)
					require.NoError(t, err)

					trailer, err := io.ReadAll(r)
					require.NoError(t, err)
					require.Equal(t, []byte(tc.trailer), trailer)
				})
			}
		})

		t.Run("error", func(t *testing.T) {
			testCases := []struct {
				input         string
				unexpectedEOF bool
				errMsg        string
			}{
				{"", true, "Header.Unmarshal: cannot peek the first 2 bytes of the line: unexpected EOF"},
				{": ", false, "Header.Unmarshal: header name is empty"},
				{"  ", false, "Header.Unmarshal: cannot read header name:"},
				{"header-name:", true, "Header.Unmarshal: cannot read header value:"},
				{"header-name:  ", true, "Header.Unmarshal: cannot read header value:"},
				{"header-name: \r\n", false, "Header.Unmarshal: header value is empty"},
				{"same: old-value\r\nsame: new-value\r\n", false, `Headers.Unmarshal: header "same" already set`},
			}

			for i, tc := range testCases {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					r := bufio.NewReader(bytes.NewReader([]byte(tc.input)))
					header := &Header{}
					err := header.Unmarshal(r)
					if tc.unexpectedEOF {
						require.ErrorIs(t, err, io.ErrUnexpectedEOF)
					}
					require.ErrorContains(t, err, tc.errMsg)
				})
			}

			t.Run("empty", func(t *testing.T) {
				t.SkipNow()
				r := bufio.NewReader(bytes.NewReader([]byte(nil)))

				var actual Header
				err := actual.Unmarshal(r)
				require.ErrorIs(t, err, io.ErrUnexpectedEOF)
				require.ErrorContains(t, err, "Header.Unmarshal: cannot peek the first")
			})
		})
	})
}
