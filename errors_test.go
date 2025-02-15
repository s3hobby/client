package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestServerSideError(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		expected := &ServerSideError{
			OperationName: "PutObject",
			Code:          "my-code",
			Message:       "my-message",
			RequestID:     "my-request-id",
			HostID:        "my-host-id",
		}

		resp := new(fasthttp.Response)
		resp.SetBody([]byte(`<Error>
			<Code>` + expected.Code + `</Code>
			<Message>` + expected.Message + `</Message>
			<RequestId>` + expected.RequestID + `</RequestId>
			<HostId>` + expected.HostID + `</HostId>
		</Error>`))

		actual, err := NewServerSideError(expected.OperationName, resp)
		require.NoError(t, err)

		actual.Response = nil

		require.Equal(t, expected, actual)
	})

	t.Run("Error", func(t *testing.T) {
		sse := &ServerSideError{
			OperationName: "PutObject",
			Code:          "my-code",
			Message:       "my-message",
			RequestID:     "my-request-id",
			HostID:        "my-host-id",
		}

		require.Equal(
			t,
			"server-side error occurred (Code:my-code) (RequestID:my-request-id) (HostID:my-host-id) when calling the PutObject operation: my-message",
			sse.Error(),
		)
	})
}
