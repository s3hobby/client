package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/s3hobby/client/pkg/utils"

	"github.com/valyala/fasthttp"
)

var _ requiredBucketInterface = (*HeadBucketInput)(nil)

type HeadBucketInput struct {
	// Bucket is mandatory
	Bucket string

	ExpectedBucketOwner *string
}

func (input *HeadBucketInput) bucket() string {
	return input.Bucket
}

func (input *HeadBucketInput) marshalHTTP(req *fasthttp.Request) error {
	req.Header.SetMethod(fasthttp.MethodHead)

	if input.ExpectedBucketOwner != nil {
		req.Header.Set("x-amz-expected-bucket-owner", *input.ExpectedBucketOwner)
	}

	return nil
}

type HeadBucketOutput struct {
	BucketRegion *string
	Metadata     *Metadata
}

func (output *HeadBucketOutput) setMetadata(v *Metadata) {
	output.Metadata = v
}

func (output *HeadBucketOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	switch resp.StatusCode() {
	case fasthttp.StatusNotFound:
		return errors.New("HeadBucket: bucket not found")
	case fasthttp.StatusForbidden:
		return errors.New("HeadBucket: fasthttp.StatusForbidden")
	case fasthttp.StatusMovedPermanently:
		return errors.New("HeadBucket: fasthttp.StatusMovedPermanently")
	case fasthttp.StatusOK:
		break
	default:
		return fmt.Errorf("HeadBucket: unexpected response: %d", resp.StatusCode())
	}

	if bucketRegion := resp.Header.Peek("x-amz-bucket-region"); bucketRegion != nil {
		output.BucketRegion = utils.ToPtr(string(bucketRegion))
	}

	return nil
}

func (c *Client) HeadBucket(ctx context.Context, input *HeadBucketInput, optFns ...func(*Options)) (*HeadBucketOutput, error) {
	in := newHandlerInput(input, "HeadBucket", c.options.With(optFns...))
	defer in.ReleaseHTTP()

	out, err := handleCall[*HeadBucketInput, HeadBucketOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
