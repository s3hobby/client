package client

import (
	"context"
	"testing"

	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestClient_HeadBucket(t *testing.T) {
	expected := &HeadBucketOutput{}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://the-bucket.s3.dev-local-1.s3hobby.local/", ctx.URI().String())
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

	resp, err := c.HeadBucket(context.Background(), &HeadBucketInput{Bucket: "the-bucket"})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Metadata)
	resp.Metadata = nil
	require.Equal(t, expected, resp)
}
