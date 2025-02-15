package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestClient_ListBuckets(t *testing.T) {
	expected := &ListBucketsOutput{
		Body: []byte("pwouet"),
	}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://s3.dev-local-1.s3hobby.local/", ctx.URI().String())

		ctx.Response.SetBody(bytes.Clone(expected.Body))
	})
	defer in.Close()

	c, err := New(&Options{
		SiginingRegion: "dev-local-1",
		EndpointHost:   "s3.dev-local-1.s3hobby.local",
		Signer:         signer.NewAnonymousSigner(),
		HTTPClient:     in.Client(),
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	resp, err := c.ListBuckets(context.Background(), &ListBucketsInput{})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Metadata)
	resp.Metadata = nil
	require.Equal(t, expected, resp)
}
