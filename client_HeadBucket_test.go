package client_test

import (
	"context"
	"testing"

	"github.com/s3hobby/client"
	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestClient_HeadBucket(t *testing.T) {
	expected := &client.HeadBucketOutput{}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://the-bucket.s3.dev-local-1.s3hobby.local/", ctx.URI().String())
	})
	defer in.Close()

	c, err := client.New(&client.Options{
		SiginingRegion: "dev-local-1",
		EndpointHost:   "s3.dev-local-1.s3hobby.local",
		Signer:         signer.NewAnonymousSigner(),
		HTTPClient:     in.Client(),
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	resp, _, err := c.HeadBucket(context.Background(), &client.HeadBucketInput{Bucket: "the-bucket"})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.Equal(t, expected, resp)
}
