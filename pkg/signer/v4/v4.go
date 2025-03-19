package v4

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/s3hobby/client/pkg/signer/utils"

	"github.com/valyala/fasthttp"
)

var emptyHash = utils.Hex(utils.SHA256Hash(nil))

type RequestChunker interface {
	Do(req *fasthttp.Request, seed, date, scope string, signingKey []byte)
}

type defaultRequestChunker struct{}

func NewDefaultRequestChunker() RequestChunker {
	return &defaultRequestChunker{}
}

func (*defaultRequestChunker) Do(req *fasthttp.Request, seed, date, scope string, signingKey []byte) {
	decodedContentLength := len(req.Body())

	var trailing string
	if actual := req.Header.Peek(HeaderContentEncoding); len(actual) > 0 {
		trailing = "," + string(actual)
	}

	req.Header.SetContentEncoding("aws-chunked" + trailing)
	req.Header.Set(HeaderXAmzDecodedContentLength, strconv.Itoa(decodedContentLength))

	const chunkSize int = 64 * 1024

	var buf bytes.Buffer
	// Grow to the approximate encoded size.
	// 1024 is for the trailing headers.
	buf.Grow(decodedContentLength + (decodedContentLength%chunkSize+1)*256)

	writeChunk := func(chunk []byte, previousSignature string) string {
		stringToSign := fmt.Sprintf(
			"AWS4-HMAC-SHA256-PAYLOAD\n%s\n%s\n%s\n%s\n%s",
			date,
			scope,
			previousSignature,
			emptyHash,
			utils.Hex(utils.SHA256Hash(chunk)),
		)

		currentSignature := utils.Hex(utils.HMAC_SHA256(signingKey, []byte(stringToSign)))

		buf.WriteString(strconv.FormatInt(int64(len(chunk)), 16))
		buf.WriteString(";chunk-signature=")
		buf.WriteString(currentSignature)
		buf.WriteString("\r\n")
		buf.Write(chunk)
		buf.WriteString("\r\n")

		return currentSignature
	}

	for chunk := range slices.Chunk(req.Body(), chunkSize) {
		seed = writeChunk(chunk, seed)
	}

	seed = writeChunk(nil, seed)

	if trailerName := req.Header.Peek(HeaderXAmzTrailer); len(trailerName) > 0 {
		trailerValue := req.Header.PeekBytes(trailerName)
		if len(trailerValue) == 0 {
			panic(fmt.Sprintf("no value for trailer %q", trailerName))
		}

		trailerBody := slices.Concat(trailerName, []byte{':'}, trailerValue, []byte{'\n'})

		stringToSign := fmt.Sprintf(
			"AWS4-HMAC-SHA256-PAYLOAD\n%s\n%s\n%s\n%s",
			date,
			scope,
			seed,
			utils.Hex(utils.SHA256Hash(trailerBody)),
		)

		currentSignature := utils.Hex(utils.HMAC_SHA256(signingKey, []byte(stringToSign)))
		fmt.Println(currentSignature)
		buf.Write(trailerBody)
		buf.WriteString("x-amz-trailer-signature:")
		buf.WriteString(currentSignature)
	}

	req.SetBody(buf.Bytes())
}

type BodySigner interface {
	ComputeContentSHA256(req *fasthttp.Request) string
	SignBody(req *fasthttp.Request, seed string)
}

type unsignedPayload struct{}

func NewUnsignedPayload() BodySigner {
	return &unsignedPayload{}
}

func (*unsignedPayload) ComputeContentSHA256(req *fasthttp.Request) string {
	if len(req.Header.Peek(HeaderXAmzTrailer)) > 0 {
		panic("STREAMING-UNSIGNED-PAYLOAD-TRAILER")
	}

	return "UNSIGNED-PAYLOAD"
}

func (*unsignedPayload) SignBody(req *fasthttp.Request, seed string) {}

type chunkedPayload struct{}

func NewChunkedPayload() BodySigner {
	return &chunkedPayload{}
}

func (*chunkedPayload) ComputeContentSHA256(req *fasthttp.Request) string {
	if len(req.Header.Peek(HeaderXAmzTrailer)) > 0 {
		panic("STREAMING-AWS4-HMAC-SHA256-PAYLOAD-TRAILER")
	}

	return "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"
}

func (*chunkedPayload) SignBody(req *fasthttp.Request, seed string) {
	req.Header.Set(HeaderXAmzDecodedContentLength, strconv.Itoa(len(req.Body())))

	var trailing string
	if actual := req.Header.Peek(HeaderContentEncoding); len(actual) > 0 {
		trailing = "," + string(actual)
	}

	req.Header.SetContentEncoding("aws-chunked" + trailing)
	req.Header.Set(HeaderXAmzDecodedContentLength, strconv.Itoa(len(req.Body())))

	panic(errors.ErrUnsupported)
}

type HeaderSigner struct {
	accessKey       string
	secretAccessKey string
	bodySigner      BodySigner
}

func NewHeaderSigner(accessKey, secretAccessKey string, bodySigner BodySigner) *HeaderSigner {
	if bodySigner == nil {
		bodySigner = NewUnsignedPayload()
	}

	return &HeaderSigner{
		accessKey:       accessKey,
		secretAccessKey: secretAccessKey,
		bodySigner:      bodySigner,
	}
}

type HeaderSigningCtx struct {
	HeaderSigner

	Req    *fasthttp.Request
	Region string
	Now    time.Time

	scope            string
	canonicalHeaders string
	signedHeaders    string
}

func (s *HeaderSigner) Sign(req *fasthttp.Request, region string, now time.Time) error {
	// AWS S3 specify the use of UTC
	now = now.UTC()

	req.Header.Set(HeaderXAmzDate, now.Format(FormatXAmzDate))
	req.Header.Set(HeaderXAmzContentSHA256, s.bodySigner.ComputeContentSHA256(req))

	ctx := &HeaderSigningCtx{
		HeaderSigner: *s,

		Req:    req,
		Region: region,
		Now:    now,
	}

	ctx.compute()

	signature := ctx.getSignature()

	req.Header.Set(
		"Authorization",
		fmt.Sprintf(
			"AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
			s.accessKey,
			ctx.scope,
			ctx.signedHeaders,
			signature,
		),
	)

	s.bodySigner.SignBody(req, signature)

	return nil
}

func (ctx *HeaderSigningCtx) compute() {
	ctx.computeScope()
	ctx.computeHeaders()
}

func (ctx *HeaderSigningCtx) computeScope() {
	ctx.scope = ctx.Now.Format(FormatYYYYMMDD) + "/" + ctx.Region + "/s3/aws4_request"
}

func (ctx *HeaderSigningCtx) computeHeaders() {
	normalized := make(map[string]string, ctx.Req.Header.Len())

	ctx.Req.Header.VisitAll(func(key, value []byte) {
		normalizedKey := utils.LowerCase(string(key))
		normalizedValue := utils.Trim(string(value))

		normalized[normalizedKey] = normalizedValue
	})

	// Ensure host header presence
	if _, exists := normalized["host"]; !exists {
		normalized["host"] = string(ctx.Req.Host())
	}

	sortedHeaders := slices.Sorted(maps.Keys(normalized))
	for _, key := range sortedHeaders {
		ctx.canonicalHeaders += key + ":" + normalized[key] + "\n"
	}

	ctx.signedHeaders = strings.Join(sortedHeaders, ";")
}

func (ctx *HeaderSigningCtx) getSignature() string {
	return utils.Hex(utils.HMAC_SHA256(ctx.getSigningKey(), []byte(ctx.getStringToSign())))
}

func (ctx *HeaderSigningCtx) getStringToSign() string {
	return "AWS4-HMAC-SHA256\n" +
		ctx.Now.Format(FormatXAmzDate) + "\n" +
		ctx.scope + "\n" +
		utils.Hex(utils.SHA256Hash([]byte(ctx.getCanonicalRequest())))
}

func (ctx *HeaderSigningCtx) getSigningKey() []byte {
	dateKey := utils.HMAC_SHA256([]byte("AWS4"+ctx.secretAccessKey), []byte(ctx.Now.Format(FormatYYYYMMDD)))
	dateRegionKey := utils.HMAC_SHA256(dateKey, []byte(ctx.Region))
	dateRegionServiceKey := utils.HMAC_SHA256(dateRegionKey, []byte("s3"))
	return utils.HMAC_SHA256(dateRegionServiceKey, []byte("aws4_request"))
}

func (ctx *HeaderSigningCtx) getCanonicalRequest() string {
	var ret strings.Builder

	ret.Write(ctx.Req.Header.Method())
	ret.WriteByte('\n')

	path := string(ctx.Req.URI().PathOriginal())
	if path == "" {
		path = "/"
	}
	ret.WriteString(utils.URIEncode(path, true))
	ret.WriteByte('\n')

	ret.WriteString(ctx.getCanonicalQueryString())
	ret.WriteByte('\n')

	ret.WriteString(ctx.canonicalHeaders)
	ret.WriteByte('\n')

	ret.WriteString(ctx.signedHeaders)
	ret.WriteByte('\n')

	ret.Write(ctx.Req.Header.Peek(HeaderXAmzContentSHA256))

	return ret.String()
}

func (ctx *HeaderSigningCtx) getCanonicalQueryString() string {
	args := ctx.Req.URI().QueryArgs()
	encoded := make(map[string]string, args.Len())

	args.VisitAll(func(key, value []byte) {
		encoded[utils.URIEncode(string(key), false)] = utils.URIEncode(string(value), false)
	})

	var ret strings.Builder

	for i, key := range slices.Sorted(maps.Keys(encoded)) {
		if i > 0 {
			ret.WriteRune('&')
		}

		ret.WriteString(key)
		ret.WriteRune('=')
		ret.WriteString(encoded[key])
	}

	return ret.String()
}
