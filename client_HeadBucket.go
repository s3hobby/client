package client

import (
	"context"
	"errors"
	"fmt"
)

type HeadBucketInput struct {
	// Bucket is mandatory
	Bucket string

	ExpectedBucketOwner *string
}

type HeadBucketOutput struct {
	BucketRegion *string
}

func (c *Client) HeadBucket(_ context.Context, input *HeadBucketInput) (*HeadBucketOutput, error) {
	if input.Bucket == "" {
		return nil, errors.New("client.Client.HeadBucket: bucket is mandatory")
	}

	return nil, fmt.Errorf("client.Client.HeadBucket: %w", errors.ErrUnsupported)
}
