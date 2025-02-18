package rawhttp

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Do(t *testing.T) {
	respBody := []byte("Hello world !\n")
	requestID := uuid.NewString()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, requestID, r.Header.Get("x-request-id"))

		// Prevent date header to be sent by the server
		w.Header()["Date"] = []string{}

		w.Header().Set("Content-Length", strconv.Itoa(len(respBody)))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		n, err := w.Write(respBody)
		if assert.NoError(t, err) {
			assert.Equal(t, len(respBody), n)
		}
	}))
	defer s.Close()

	c := &Client{
		Dial: func(ctx context.Context, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			return dialer.DialContext(ctx, "tcp", addr)
		},
	}

	req := &Message{
		StartLine: StartLine{MethodGet, "/", HTTPVersion1_1},
		Header: Header{data: map[string]string{
			"x-request-id":   requestID,
			"host":           s.Listener.Addr().String(),
			"content-length": "0",
		}},
	}

	t.Log("Will connect to:", s.Listener.Addr().String())

	resp, err := c.Do(t.Context(), s.Listener.Addr().String(), req)
	require.NoError(t, err)

	expected := &Message{
		StartLine: StartLine{"HTTP/1.1", "200", "OK"},
		Header: Header{data: map[string]string{
			"content-length": strconv.Itoa(len(respBody)),
			"content-type":   "text/plain",
		}},
	}
	require.Equal(t, expected, resp)
}
