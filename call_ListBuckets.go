package client

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/s3hobby/client/types"

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

	args := req.URI().QueryArgs()
	setQuery(args, QueryBucketRegion, input.BucketRegion)
	setQuery(args, QueryContinuationToken, input.ContinuationToken)
	setQuery(args, QueryMaxBuckets, input.MaxBuckets)
	setQuery(args, QueryPrefix, input.Prefix)

	return nil
}

type ListBucketsOutput struct {
	Payload *types.ListAllMyBucketsResult
}

func (output *ListBucketsOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	switch resp.StatusCode() {
	case fasthttp.StatusForbidden:
		return errors.New("ListBuckets: fasthttp.StatusForbidden")
	case fasthttp.StatusOK:
		var payload types.ListAllMyBucketsResult
		if err := xml.Unmarshal(resp.Body(), &payload); err != nil {
			return fmt.Errorf("ListBuckets: cannot parse response body: %w", err)
		}
		output.Payload = &payload
		return nil
	default:
		return fmt.Errorf("ListBuckets: unexpected response: %d", resp.StatusCode())
	}
}

func (c *Client) ListBuckets(ctx context.Context, input *ListBucketsInput, optFns ...func(*Options)) (*ListBucketsOutput, error) {
	in := &handlerInput[*ListBucketsInput]{
		Options:           c.options.With(optFns...),
		SuccessStatusCode: fasthttp.StatusOK,
		CallInput:         input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[*ListBucketsInput, *ListBucketsOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
