package v4

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/s3hobby/client/pkg/signer/utils"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestHeaderSigningCtx(t *testing.T) {
	testCases := []struct {
		name                     string
		req                      func(*fasthttp.Request)
		expectedCanonicalRequest string
		expectedStringToSign     string
		expectedSignature        string
		expectedAuthorization    string
	}{
		{
			name: "GET Object",
			req: func(req *fasthttp.Request) {
				req.SetRequestURI("http://examplebucket.s3.amazonaws.com/test.txt")
				req.Header.SetMethod("GET")
				req.Header.SetByteRange(0, 9)
				req.Header.Set("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
				req.Header.Set("x-amz-date", "20130524T000000Z")
			},
			expectedCanonicalRequest: `GET
/test.txt

host:examplebucket.s3.amazonaws.com
range:bytes=0-9
x-amz-content-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
x-amz-date:20130524T000000Z

host;range;x-amz-content-sha256;x-amz-date
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
			expectedStringToSign: `AWS4-HMAC-SHA256
20130524T000000Z
20130524/us-east-1/s3/aws4_request
7344ae5b7ee6c3e7e6b0fe0640412a37625d1fbfff95c48bbb2dc43964946972`,
			expectedSignature:     "f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41",
			expectedAuthorization: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41",
		},
		{
			name: "PUT Object",
			req: func(req *fasthttp.Request) {
				req.Header.SetMethod("PUT")
				req.SetRequestURI("http://examplebucket.s3.amazonaws.com/test$file.text")
				req.Header.Set("Date", "Fri, 24 May 2013 00:00:00 GMT")
				req.Header.Set("x-amz-date", "20130524T000000Z")
				req.Header.Set("x-amz-storage-class", "REDUCED_REDUNDANCY")
				req.Header.Set("x-amz-content-sha256", "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072")
			},
			expectedCanonicalRequest: `PUT
/test%24file.text

date:Fri, 24 May 2013 00:00:00 GMT
host:examplebucket.s3.amazonaws.com
x-amz-content-sha256:44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072
x-amz-date:20130524T000000Z
x-amz-storage-class:REDUCED_REDUNDANCY

date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class
44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072`,
			expectedStringToSign: `AWS4-HMAC-SHA256
20130524T000000Z
20130524/us-east-1/s3/aws4_request
9e0e90d9c76de8fa5b200d8c849cd5b8dc7a3be3951ddb7f6a76b4158342019d`,
			expectedSignature:     "98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd",
			expectedAuthorization: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class,Signature=98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd",
		},
		{
			name: "GET Bucket Lifecycle",
			req: func(req *fasthttp.Request) {
				req.Header.SetMethod("GET")
				req.SetRequestURI("http://examplebucket.s3.amazonaws.com?lifecycle")
				req.Header.Set("x-amz-date", "20130524T000000Z")
				req.Header.Set("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
			},
			expectedCanonicalRequest: `GET
/
lifecycle=
host:examplebucket.s3.amazonaws.com
x-amz-content-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
x-amz-date:20130524T000000Z

host;x-amz-content-sha256;x-amz-date
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
			expectedStringToSign: `AWS4-HMAC-SHA256
20130524T000000Z
20130524/us-east-1/s3/aws4_request
9766c798316ff2757b517bc739a67f6213b4ab36dd5da2f94eaebf79c77395ca`,
			expectedSignature:     "fea454ca298b7da1c68078a5d1bdbfbbe0d65c699e0f91ac7a200a0136783543",
			expectedAuthorization: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;x-amz-content-sha256;x-amz-date,Signature=fea454ca298b7da1c68078a5d1bdbfbbe0d65c699e0f91ac7a200a0136783543",
		},
		{
			name: "Get Bucket (List Objects)",
			req: func(req *fasthttp.Request) {
				req.Header.SetMethod("GET")
				req.SetRequestURI("http://examplebucket.s3.amazonaws.com?max-keys=2&prefix=J")
				req.Header.Set("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
				req.Header.Set("x-amz-date", "20130524T000000Z")
			},
			expectedCanonicalRequest: `GET
/
max-keys=2&prefix=J
host:examplebucket.s3.amazonaws.com
x-amz-content-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
x-amz-date:20130524T000000Z

host;x-amz-content-sha256;x-amz-date
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
			expectedStringToSign: `AWS4-HMAC-SHA256
20130524T000000Z
20130524/us-east-1/s3/aws4_request
df57d21db20da04d7fa30298dd4488ba3a2b47ca3a489c74750e0f1e7df1b9b7`,
			expectedSignature:     `34b48302e7b5fa45bde8084f4b7868a86f0a534bc59db6670ed5711ef69dc6f7`,
			expectedAuthorization: `AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;x-amz-content-sha256;x-amz-date,Signature=34b48302e7b5fa45bde8084f4b7868a86f0a534bc59db6670ed5711ef69dc6f7`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req fasthttp.Request
			tc.req(&req)

			ctx := &HeaderSigningCtx{
				HeaderSigner: HeaderSigner{
					accessKey:       "AKIAIOSFODNN7EXAMPLE",
					secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
				Req:    &req,
				Region: "us-east-1",
				Now:    time.Date(2013, time.May, 24, 0, 0, 0, 0, time.UTC),
			}

			ctx.compute()

			require.Equal(t, "20130524/us-east-1/s3/aws4_request", ctx.scope)
			require.Equal(t, tc.expectedCanonicalRequest, ctx.getCanonicalRequest())
			require.Equal(t, tc.expectedStringToSign, ctx.getStringToSign())
			require.Equal(t, tc.expectedSignature, ctx.getSignature())
		})
	}
}

func TestDefaultRequestChunker(t *testing.T) {
	var req fasthttp.Request
	req.SetRequestURI("http://s3.amazonaws.com/examplebucket/chunkObject.txt")
	req.Header.SetMethod("PUT")
	req.Header.Set("x-amz-date", "20130524T000000Z")
	req.Header.Set("x-amz-storage-class", "REDUCED_REDUNDANCY")
	req.SetBody(slices.Repeat([]byte{'a'}, 65*1024))

	signingKey := utils.HMAC_SHA256(
		utils.HMAC_SHA256(
			utils.HMAC_SHA256(
				utils.HMAC_SHA256([]byte("AWS4"+"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"), []byte("20130524")),
				[]byte("us-east-1")),
			[]byte("s3")),
		[]byte("aws4_request"),
	)

	chunker := NewDefaultRequestChunker()
	chunker.Do(
		&req,
		"4f232c4386841ef735655705268965c44a0e4690baa4adea153f7db9fa80a0a9",
		"20130524T000000Z",
		"20130524/us-east-1/s3/aws4_request",
		signingKey,
	)

	expectedBody := "10000;chunk-signature=ad80c730a21e5b8d04586a2213dd63b9a0e99e0e2307b0ade35a65485a288648\r\n" +
		string(slices.Repeat([]byte{'a'}, 64*1024)) + "\r\n" +
		"400;chunk-signature=0055627c9e194cb4542bae2aa5492e3c1575bbb81b612b7d234b86a503ef5497" + "\r\n" +
		string(slices.Repeat([]byte{'a'}, 1024)) + "\r\n" +
		"0;chunk-signature=b6c6ea8a5354eaf15b3cb7646744f4275b71ea724fed81ceb9323e279d449df9" + "\r\n" + "\r\n"

	require.Equal(t, expectedBody, string(req.Body()))
}

func TestDefaultRequestChunker2(t *testing.T) {
	var req fasthttp.Request
	req.SetRequestURI("http://s3.amazonaws.com/examplebucket/chunkObject.txt")
	req.Header.SetNoDefaultContentType(true)
	req.Header.SetMethod("PUT")
	req.Header.Set("x-amz-date", "20130524T000000Z")
	req.Header.Set("x-amz-storage-class", "REDUCED_REDUNDANCY")
	req.Header.Set("x-amz-trailer", "x-amz-checksum-crc32c")
	req.Header.Set("x-amz-checksum-crc32c", "sOO8/Q==")
	req.SetBody(slices.Repeat([]byte{'a'}, 65*1024))

	signingKey := utils.HMAC_SHA256(
		utils.HMAC_SHA256(
			utils.HMAC_SHA256(
				utils.HMAC_SHA256([]byte("AWS4"+"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"), []byte("20130524")),
				[]byte("us-east-1")),
			[]byte("s3")),
		[]byte("aws4_request"),
	)

	chunker := NewDefaultRequestChunker()
	chunker.Do(
		&req,
		"106e2a8a18243abcf37539882f36619c00e2dfc72633413f02d3b74544bfeb8e",
		"20130524T000000Z",
		"20130524/us-east-1/s3/aws4_request",
		signingKey,
	)

	expectedBody := "10000;chunk-signature=b474d8862b1487a5145d686f57f013e54db672cee1c953b3010fb58501ef5aa2\r\n" +
		string(slices.Repeat([]byte{'a'}, 64*1024)) + "\r\n" +
		"400;chunk-signature=1c1344b170168f8e65b41376b44b20fe354e373826ccbbe2c1d40a8cae51e5c7" + "\r\n" +
		string(slices.Repeat([]byte{'a'}, 1024)) + "\r\n" +
		"0;chunk-signature=2ca2aba2005185cf7159c6277faf83795951dd77a3a99e6e65d5c9f85863f992" + "\r\n" + "\r\n" +
		"x-amz-checksum-crc32c:sOO8/Q==" + "\n" +
		"x-amz-trailer-signature:63bddb248ad2590c92712055f51b8e78ab024eead08276b24f010b0efd74843f"

	_, err := req.WriteTo(os.Stdout)
	require.NoError(t, err)

	require.Equal(t, expectedBody, string(req.Body()))
}
