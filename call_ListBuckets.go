package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
)

type ListBucketsInput struct {
	BucketRegion      *string
	ContinuationToken *string
	MaxBuckets        *string
	Prefix            *string
}

func (input *ListBucketsInput) marshalHTTP(req *fasthttp.Request) error {
	req.Header.SetMethod(fasthttp.MethodGet)

	if input.BucketRegion != nil {
		req.URI().QueryArgs().Set("bucket-region", *input.BucketRegion)
	}

	if input.ContinuationToken != nil {
		req.URI().QueryArgs().Set("continuation-token", *input.ContinuationToken)
	}

	if input.MaxBuckets != nil {
		req.URI().QueryArgs().Set("max-buckets", *input.MaxBuckets)
	}

	if input.Prefix != nil {
		req.URI().QueryArgs().Set("prefix", *input.Prefix)
	}

	return nil
}

type ListBucketsOutput struct {
	// TODO Correctly implements fields
	Body []byte

	Metadata *Metadata
}

func (output *ListBucketsOutput) setMetadata(v *Metadata) {
	output.Metadata = v
}

func (output *ListBucketsOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	switch resp.StatusCode() {
	case fasthttp.StatusForbidden:
		return errors.New("ListBuckets: fasthttp.StatusForbidden")
	case fasthttp.StatusOK:
		output.Body = bytes.Clone(resp.Body())
	default:
		return fmt.Errorf("ListBuckets: unexpected response: %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) ListBuckets(ctx context.Context, input *ListBucketsInput, optFns ...func(*Options)) (*ListBucketsOutput, error) {
	in := &handlerInput[*ListBucketsInput]{
		OperationName:     "ListBuckets",
		Options:           c.options.With(optFns...),
		SuccessStatusCode: fasthttp.StatusOK,
		CallInput:         input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[*ListBucketsInput, ListBucketsOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
