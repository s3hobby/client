package client_test

import (
	"context"
	"testing"

	"github.com/s3hobby/client"
	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"
	"github.com/s3hobby/client/pkg/utils"
	"github.com/s3hobby/client/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestClient_ListBuckets(t *testing.T) {
	expected := &client.ListBucketsOutput{
		Payload: &types.ListAllMyBucketsResult{
			Buckets: []types.Bucket{{
				BucketRegion: utils.ToPtr("dev-local-1"),
				Name:         utils.ToPtr("my-bucket"),
			}},
		},
	}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://s3.dev-local-1.s3hobby.local/", ctx.URI().String())

		ctx.Response.SetBody([]byte(`<ListAllMyBucketsResult>
	<Buckets>
		<Bucket>
			<BucketRegion>dev-local-1</BucketRegion>
			<Name>my-bucket</Name>
		</Bucket>
	</Buckets>
</ListAllMyBucketsResult>`))
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

	resp, _, err := c.ListBuckets(context.Background(), &client.ListBucketsInput{})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.Equal(t, expected, resp)
}
