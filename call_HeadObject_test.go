package client

import (
	"context"
	"testing"

	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"
	"github.com/s3hobby/client/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestClient_HeadObject(t *testing.T) {
	expected := &HeadObjectOutput{
		ContentLength: 1234,
		ETag:          utils.ToPtr("my-etag"),
		LastModified:  utils.ToPtr("last-modified"),
	}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://the-bucket.s3.dev-local-1.s3hobby.local/the-key", ctx.URI().String())

		ctx.Response.Header.SetContentLength(expected.ContentLength)
		ctx.Response.Header.Set(fasthttp.HeaderETag, *expected.ETag)
		ctx.Response.Header.Set(fasthttp.HeaderLastModified, *expected.LastModified)
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

	resp, err := c.HeadObject(context.Background(), &HeadObjectInput{
		Bucket: "the-bucket",
		Key:    "the-key",
	})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Metadata)
	resp.Metadata = nil
	require.Equal(t, expected, resp)
}
