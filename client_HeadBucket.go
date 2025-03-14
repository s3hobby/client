package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
)

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

	setHeader(&req.Header, HeaderXAmzExpectedBucketOwner, input.ExpectedBucketOwner)

	return nil
}

type HeadBucketOutput struct {
	AccessPointAlias *string
	BucketRegion     *string
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

	output.BucketRegion = extractHeader(&resp.Header, HeaderXAmzBucketRegion)
	output.AccessPointAlias = extractHeader(&resp.Header, HeaderXamzAccessPointAlias)

	return nil
}

func (c *Client) HeadBucket(ctx context.Context, input *HeadBucketInput, optFns ...func(*Options)) (*HeadBucketOutput, error) {
	in := &handlerInput[*HeadBucketInput]{
		Options:           c.options.With(optFns...),
		SuccessStatusCode: fasthttp.StatusOK,
		CallInput:         input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[*HeadBucketInput, *HeadBucketOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
