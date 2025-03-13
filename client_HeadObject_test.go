package client_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/s3hobby/client"
	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"
	"github.com/s3hobby/client/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestHeadObject(t *testing.T) {
	// t.Run("marshal", func(t *testing.T) {
	// 	hoi := &client.HeadObjectInput{
	// 		Bucket: newBucket(),
	// 		Key:    newObjectKey(),

	// 		PartNumber:                 newURLQueryValue(),
	// 		ResponseCacheControl:       newURLQueryValue(),
	// 		ResponseContentDisposition: newURLQueryValue(),
	// 		ResponseContentEncoding:    newURLQueryValue(),
	// 		ResponseContentLanguage:    newURLQueryValue(),
	// 		ResponseContentType:        newURLQueryValue(),
	// 		ResponseExpires:            newURLQueryValue(),
	// 		VersionId:                  newURLQueryValue(),

	// 		IfMatch:              newHeaderValue(),
	// 		IfModifiedSince:      newHeaderValue(),
	// 		IfNoneMatch:          newHeaderValue(),
	// 		IfUnmodifiedSince:    newHeaderValue(),
	// 		Range:                newHeaderValue(),
	// 		SSECustomerAlgorithm: newHeaderValue(),
	// 		SSECustomerKey:       newHeaderValue(),
	// 		SSECustomerKeyMD5:    newHeaderValue(),
	// 		RequestPayer:         newHeaderValue(),
	// 		ExpectedBucketOwner:  newHeaderValue(),
	// 		ChecksumMode:         newHeaderValue(),
	// 	}

	// 	expectedQueries := map[string]string{
	// 		"partNumber":                   *hoi.PartNumber,
	// 		"response-cache-control":       *hoi.ResponseCacheControl,
	// 		"response-content-disposition": *hoi.ResponseContentDisposition,
	// 		"response-content-encoding":    *hoi.ResponseContentEncoding,
	// 		"response-content-language":    *hoi.ResponseContentLanguage,
	// 		"response-content-type":        *hoi.ResponseContentType,
	// 		"response-expires":             *hoi.ResponseExpires,
	// 		"versionId":                    *hoi.VersionId,
	// 	}

	// 	expectedHeaders := map[string]string{
	// 		"If-Match":            *hoi.IfMatch,
	// 		"If-Modified-Since":   *hoi.IfModifiedSince,
	// 		"If-None-Match":       *hoi.IfNoneMatch,
	// 		"If-Unmodified-Since": *hoi.IfUnmodifiedSince,
	// 		"Range":               *hoi.Range,
	// 		"x-amz-server-side-encryption-customer-algorithm": *hoi.SSECustomerAlgorithm,
	// 		"x-amz-server-side-encryption-customer-key":       *hoi.SSECustomerKey,
	// 		"x-amz-server-side-encryption-customer-key-MD5":   *hoi.SSECustomerKeyMD5,
	// 		"x-amz-request-payer":                             *hoi.RequestPayer,
	// 		"x-amz-expected-bucket-owner":                     *hoi.ExpectedBucketOwner,
	// 		"x-amz-checksum-mode":                             *hoi.ChecksumMode,
	// 	}

	// 	////////////////////////////////////////////////////////////////////////////////
	// 	params := client.EndpointParameters{
	// 		Bucket:       hoi.Bucket,
	// 		Key:          hoi.Key,
	// 		Host:         "s3.dev-1.s3hobby.local",
	// 		UseSSL:       true,
	// 		UsePathStyle: false,
	// 	}

	// 	endpoint, err := client.DefaultEndpointResolver.ResolveEndpoint(t.Context(), params)
	// 	require.NoError(t, err)

	// 	var req fasthttp.Request
	// 	req.SetRequestURI(endpoint.URL)
	// 	// require.NoError(t, hoi.MarshalHTTP(&req))
	// 	////////////////////////////////////////////////////////////////////////////////

	// 	require.Equal(t, fasthttp.MethodHead, string(req.Header.Method()))
	// 	require.Equal(t, hoi.Bucket+".s3.dev-1.s3hobby.local", string(req.Host()))
	// 	require.Equal(t, "/"+hoi.Key, string(req.URI().Path()))

	// 	require.Equal(t, len(expectedQueries), req.URI().QueryArgs().Len())
	// 	req.URI().QueryArgs().VisitAll(func(key, actual []byte) {
	// 		require.Contains(t, expectedQueries, string(key))
	// 		require.Equal(t, expectedQueries[string(key)], string(actual))
	// 	})

	// 	require.Equal(t, len(expectedHeaders), req.Header.Len())
	// 	for key, value := range expectedHeaders {
	// 		require.Equal(t, value, string(req.Header.Peek(key)))
	// 	}
	// })

	t.Run("unmarshal", func(t *testing.T) {

	})

	const contentLength = 1234

	expected := &client.HeadObjectOutput{
		ContentLength: utils.ToPtr(strconv.Itoa(contentLength)),
		ETag:          utils.ToPtr("my-etag"),
		LastModified:  utils.ToPtr("last-modified"),
	}

	in := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "http://the-bucket.s3.dev-local-1.s3hobby.local/the-key", ctx.URI().String())

		ctx.Response.Header.SetContentLength(contentLength)
		ctx.Response.Header.Set(fasthttp.HeaderETag, *expected.ETag)
		ctx.Response.Header.Set(fasthttp.HeaderLastModified, *expected.LastModified)
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

	resp, _, err := c.HeadObject(context.Background(), &client.HeadObjectInput{
		Bucket: "the-bucket",
		Key:    "the-key",
	})
	require.NoError(t, err)

	require.NotNil(t, resp)
	require.Equal(t, expected, resp)
}
