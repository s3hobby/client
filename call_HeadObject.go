package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/s3hobby/client/pkg/utils"

	"github.com/valyala/fasthttp"
)

var _ requiredBucketKeyInterface = (*HeadObjectInput)(nil)

type HeadObjectInput struct {
	// Bucket is mandatory
	Bucket string

	// Key is mandatory
	Key string

	PartNumber                 *string
	ResponseCacheControl       *string
	ResponseContentDisposition *string
	ResponseContentEncoding    *string
	ResponseContentLanguage    *string
	ResponseContentType        *string
	ResponseExpires            *string
	VersionId                  *string

	IfMatch              *string
	IfModifiedSince      *string
	IfNoneMatch          *string
	IfUnmodifiedSince    *string
	Range                *string
	SSECustomerAlgorithm *string
	SSECustomerKey       *string
	SSECustomerKeyMD5    *string
	RequestPayer         *string
	ExpectedBucketOwner  *string
	ChecksumMode         *string
}

func (input *HeadObjectInput) bucket() string {
	return input.Bucket
}

func (input *HeadObjectInput) key() string {
	return input.Key
}

func (input *HeadObjectInput) marshalHTTP(req *fasthttp.Request) error {
	req.Header.SetMethod(fasthttp.MethodHead)

	for argName, value := range map[string]*string{
		"partNumber":                   input.PartNumber,
		"response-cache-control":       input.ResponseCacheControl,
		"response-content-disposition": input.ResponseContentDisposition,
		"response-content-encoding":    input.ResponseContentEncoding,
		"response-content-language":    input.ResponseContentLanguage,
		"response-content-type":        input.ResponseContentType,
		"response-expires":             input.ResponseExpires,
		"versionId":                    input.VersionId,
	} {
		if value != nil {
			req.URI().QueryArgs().Set(argName, *value)
		}
	}

	for argName, value := range map[string]*string{
		"If-Match":            input.IfMatch,
		"If-Modified-Since":   input.IfModifiedSince,
		"If-None-Match":       input.IfNoneMatch,
		"If-Unmodified-Since": input.IfUnmodifiedSince,
		"Range":               input.Range,
		"x-amz-server-side-encryption-customer-algorithm": input.SSECustomerAlgorithm,
		"x-amz-server-side-encryption-customer-key":       input.SSECustomerKey,
		"x-amz-server-side-encryption-customer-key-MD5":   input.SSECustomerKeyMD5,
		"x-amz-request-payer":                             input.RequestPayer,
		"x-amz-expected-bucket-owner":                     input.ExpectedBucketOwner,
		"x-amz-checksum-mode":                             input.ChecksumMode,
	} {
		if value != nil {
			req.Header.Set(argName, *value)
		}
	}

	return nil
}

type HeadObjectOutput struct {
	ContentLength int
	LastModified  *string
	ETag          *string

	// TODO Implements fields

	Metadata *Metadata
}

func (output *HeadObjectOutput) setMetadata(v *Metadata) {
	output.Metadata = v
}

func (output *HeadObjectOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	switch resp.StatusCode() {
	case fasthttp.StatusNotFound:
		return errors.New("HeadObject: bucket not found")
	case fasthttp.StatusForbidden:
		return errors.New("HeadObject: fasthttp.StatusForbidden")
	case fasthttp.StatusMovedPermanently:
		return errors.New("HeadObject: fasthttp.StatusMovedPermanently")
	case fasthttp.StatusOK:
		break
	default:
		return fmt.Errorf("HeadObject: unexpected response: %d", resp.StatusCode())
	}

	output.ContentLength = resp.Header.ContentLength()
	output.LastModified = utils.ToPtr(string(resp.Header.Peek(fasthttp.HeaderLastModified)))
	output.ETag = utils.ToPtr(string(resp.Header.Peek(fasthttp.HeaderETag)))

	return nil
}

func (c *Client) HeadObject(ctx context.Context, input *HeadObjectInput, optFns ...func(*Options)) (*HeadObjectOutput, error) {
	in := &handlerInput[*HeadObjectInput]{
		OperationName:     "HeadObject",
		Options:           c.options.With(optFns...),
		SuccessStatusCode: fasthttp.StatusOK,
		CallInput:         input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[*HeadObjectInput, HeadObjectOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
