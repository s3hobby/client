package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_HeadBucket(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	require.NotNil(t, c)

	t.Run("mandatory bucket", func(t *testing.T) {
		_, err := c.HeadBucket(context.Background(), &HeadBucketInput{})
		require.EqualError(t, err, "client.Client.HeadBucket: bucket is mandatory")
	})

	t.Run("simple", func(t *testing.T) {
		t.Skip("Not yet implemented")
		_, err := c.HeadBucket(context.Background(), &HeadBucketInput{Bucket: "the-bucket"})
		require.NoError(t, err)
	})
}
