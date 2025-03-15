package client

import (
	"context"
	"fmt"

	"github.com/valyala/fasthttp"
)

type PutObjectInput struct {
	// Bucket is mandatory
	Bucket string

	// Key is mandatory
	Key string

	Body []byte

	CacheControl       *string
	ContentDisposition *string
	ContentEncoding    *string
	ContentLanguage    *string
	ContentMD5         *string
	ContentType        *string
	Expires            *string
	IfMatch            *string
	IfNoneMatch        *string

	ACL                       *string
	ChecksumCRC32             *string
	ChecksumCRC32C            *string
	ChecksumCRC64NVME         *string
	ChecksumSHA1              *string
	ChecksumSHA256            *string
	ExpectedBucketOwner       *string
	GrantFullControl          *string
	GrantReadACP              *string
	GrantRead                 *string
	GrantWriteACP             *string
	ObjectLockLegalHoldStatus *string
	ObjectLockMode            *string
	ObjectLockRetainUntilDate *string
	RequestPayer              *string
	ChecksumAlgorithm         *string
	SSEKMSKeyId               *string
	BucketKeyEnabled          *string
	SSEKMSEncryptionContext   *string
	SSECustomerAlgorithm      *string
	SSECustomerKeyMD5         *string
	SSECustomerKey            *string
	ServerSideEncryption      *string
	StorageClass              *string
	Tagging                   *string
	WebsiteRedirectLocation   *string
	WriteOffsetBytes          *string
}

func (input *PutObjectInput) bucket() string {
	return input.Bucket
}

func (input *PutObjectInput) key() string {
	return input.Key
}

func (input *PutObjectInput) marshalHTTP(req *fasthttp.Request) error {
	req.Header.SetMethod(fasthttp.MethodPut)

	req.ResetBody()
	if input.Body != nil {
		req.SetBody(input.Body)
	}

	setHeader(&req.Header, HeaderCacheControl, input.CacheControl)
	setHeader(&req.Header, HeaderContentDisposition, input.ContentDisposition)
	setHeader(&req.Header, HeaderContentEncoding, input.ContentEncoding)
	setHeader(&req.Header, HeaderContentLanguage, input.ContentLanguage)
	setHeader(&req.Header, HeaderContentMD5, input.ContentMD5)
	setHeader(&req.Header, HeaderContentType, input.ContentType)
	setHeader(&req.Header, HeaderExpires, input.Expires)
	setHeader(&req.Header, HeaderIfMatch, input.IfMatch)
	setHeader(&req.Header, HeaderIfNoneMatch, input.IfNoneMatch)

	setHeader(&req.Header, HeaderXAmzACL, input.ACL)
	setHeader(&req.Header, HeaderXAmzChecksumCRC32, input.ChecksumCRC32)
	setHeader(&req.Header, HeaderXAmzChecksumCRC32C, input.ChecksumCRC32C)
	setHeader(&req.Header, HeaderXAmzChecksumCRC64NVME, input.ChecksumCRC64NVME)
	setHeader(&req.Header, HeaderXAmzChecksumSHA1, input.ChecksumSHA1)
	setHeader(&req.Header, HeaderXAmzChecksumSHA256, input.ChecksumSHA256)
	setHeader(&req.Header, HeaderXAmzExpectedBucketOwner, input.ExpectedBucketOwner)
	setHeader(&req.Header, HeaderXAmzGrantFullControl, input.GrantFullControl)
	setHeader(&req.Header, HeaderXAmzGrantReadACP, input.GrantReadACP)
	setHeader(&req.Header, HeaderXAmzGrantRead, input.GrantRead)
	setHeader(&req.Header, HeaderXAmzGrantWriteACP, input.GrantWriteACP)
	setHeader(&req.Header, HeaderXAmzObjectLockLegalHoldStatus, input.ObjectLockLegalHoldStatus)
	setHeader(&req.Header, HeaderXAmzObjectLockMode, input.ObjectLockMode)
	setHeader(&req.Header, HeaderXAmzObjectLockRetainUntilDate, input.ObjectLockRetainUntilDate)
	setHeader(&req.Header, HeaderXAmzRequestPayer, input.RequestPayer)
	setHeader(&req.Header, HeaderXAmzChecksumAlgorithm, input.ChecksumAlgorithm)
	setHeader(&req.Header, HeaderXAmzSSEKMSKeyId, input.SSEKMSKeyId)
	setHeader(&req.Header, HeaderXAmzBucketKeyEnabled, input.BucketKeyEnabled)
	setHeader(&req.Header, HeaderXAmzSSEKMSEncryptionContext, input.SSEKMSEncryptionContext)
	setHeader(&req.Header, HeaderXAmzSSECustomerAlgorithm, input.SSECustomerAlgorithm)
	setHeader(&req.Header, HeaderXAmzSSECustomerKeyMD5, input.SSECustomerKeyMD5)
	setHeader(&req.Header, HeaderXAmzSSECustomerKey, input.SSECustomerKey)
	setHeader(&req.Header, HeaderXAmzServerSideEncryption, input.ServerSideEncryption)
	setHeader(&req.Header, HeaderXAmzStorageClass, input.StorageClass)
	setHeader(&req.Header, HeaderXAmzTagging, input.Tagging)
	setHeader(&req.Header, HeaderXAmzWebsiteRedirectLocation, input.WebsiteRedirectLocation)
	setHeader(&req.Header, HeaderXAmzWriteOffsetBytes, input.WriteOffsetBytes)

	return nil
}

type PutObjectOutput struct {
	ETag                    *string
	ChecksumCRC32           *string
	ChecksumCRC32C          *string
	ChecksumCRC64NVME       *string
	ChecksumSHA1            *string
	ChecksumSHA256          *string
	ChecksumType            *string
	Expiration              *string
	Size                    *string
	RequestCharged          *string
	SSEKMSKeyId             *string
	BucketKeyEnabled        *string
	SSEKMSEncryptionContext *string
	SSECustomerAlgorithm    *string
	SSECustomerKeyMD5       *string
	ServerSideEncryption    *string
	VersionId               *string
}

func (output *PutObjectOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("PutObject: unexpected response: %d", resp.StatusCode())
	}

	output.ETag = extractHeader(&resp.Header, HeaderETag)
	output.ChecksumCRC32 = extractHeader(&resp.Header, HeaderXAmzChecksumCRC32)
	output.ChecksumCRC32C = extractHeader(&resp.Header, HeaderXAmzChecksumCRC32C)
	output.ChecksumCRC64NVME = extractHeader(&resp.Header, HeaderXAmzChecksumCRC64NVME)
	output.ChecksumSHA1 = extractHeader(&resp.Header, HeaderXAmzChecksumSHA1)
	output.ChecksumSHA256 = extractHeader(&resp.Header, HeaderXAmzChecksumSHA256)
	output.ChecksumType = extractHeader(&resp.Header, HeaderXAmzChecksumType)
	output.Expiration = extractHeader(&resp.Header, HeaderXAmzExpiration)
	output.Size = extractHeader(&resp.Header, HeaderXAmzSize)
	output.RequestCharged = extractHeader(&resp.Header, HeaderXAmzRequestCharged)
	output.SSEKMSKeyId = extractHeader(&resp.Header, HeaderXAmzSSEKMSKeyId)
	output.BucketKeyEnabled = extractHeader(&resp.Header, HeaderXAmzBucketKeyEnabled)
	output.SSEKMSEncryptionContext = extractHeader(&resp.Header, HeaderXAmzSSEKMSEncryptionContext)
	output.SSECustomerAlgorithm = extractHeader(&resp.Header, HeaderXAmzSSECustomerAlgorithm)
	output.SSECustomerKeyMD5 = extractHeader(&resp.Header, HeaderXAmzSSECustomerKeyMD5)
	output.ServerSideEncryption = extractHeader(&resp.Header, HeaderXAmzServerSideEncryption)
	output.VersionId = extractHeader(&resp.Header, HeaderXAmzVersionId)

	return nil
}

func (c *Client) PutObject(ctx context.Context, input *PutObjectInput, optFns ...func(*Options)) (*PutObjectOutput, error) {
	in := &handlerInput[*PutObjectInput]{
		Options:           c.options.With(optFns...),
		SuccessStatusCode: fasthttp.StatusOK,
		CallInput:         input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[*PutObjectInput, *PutObjectOutput](ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
