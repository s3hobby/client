# Operations

Region: aws-global

- Initialize
  - RegisterServiceMetadata: `github.com/aws/aws-sdk-go-v2/aws/middleware.RegisterServiceMetadata`

      ```golang
      SetServiceID: middleware.WithStackValue(ctx, serviceIDKey{}, value)
      SetSigningName: middleware.WithStackValue(ctx, signingNameKey{}, value)
      setRegion: middleware.WithStackValue(ctx, regionKey{}, value)
      setOperationName: middleware.WithStackValue(ctx, operationNameKey{}, value)
      ```

  - SetLogger: `github.com/aws/smithy-go/middleware.setLogger`
  - OperationInputValidation: `github.com/aws/aws-sdk-go-v2/service/s3.validateOpCreateBucket`
- Serialize
  - putBucketContext: `github.com/aws/aws-sdk-go-v2/service/s3.putBucketContextMiddleware`

      ```golang
      SetBucket: middleware.WithStackValue(ctx, bucketKey{}, bucket)
      ```

  - setOperationInput: `github.com/aws/aws-sdk-go-v2/service/s3.setOperationInputMiddleware`

      ```golang
      setOperationInput: middleware.WithStackValue(ctx, operationInputKey{}, input)
      ```

  - serializeImmutableHostnameBucket: `github.com/aws/aws-sdk-go-v2/service/s3.serializeImmutableHostnameBucketMiddleware`
      => BOF
  - OperationSerializer: `github.com/aws/aws-sdk-go-v2/service/s3.awsRestxml_serializeOpCreateBucket`
  - isExpressUserAgent: `github.com/aws/aws-sdk-go-v2/service/s3.isExpressUserAgent`
- Build
  - ClientRequestID: `github.com/aws/aws-sdk-go-v2/aws/middleware.ClientRequestID`

      ```golang
      req.Header["Amz-Sdk-Invocation-Id"] = uuid.NewString()
      ```

  - ComputeContentLength: `github.com/aws/smithy-go/transport/http.ComputeContentLength`

      ```golang
      req.Header["Content-Length"] = smithyReq.StreamLength()
      ```

  - UserAgent: `github.com/aws/aws-sdk-go-v2/aws/middleware.RequestUserAgent`

      ```golang
      req.Header["User-Agent"] = "aws-sdk-go-v2/1.32.5 os/macos"
      ```

- Finalize
  - DisableAcceptEncodingGzip: `github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding.DisableGzip`

      ```golang
      req.Header["Accept-Encoding"] = "identity"
      ```

  - ResolveAuthScheme: `github.com/aws/aws-sdk-go-v2/service/s3.resolveAuthSchemeMiddleware`
  - GetIdentity: `github.com/aws/aws-sdk-go-v2/service/s3.getIdentityMiddleware`
  - ResolveEndpointV2: `github.com/aws/aws-sdk-go-v2/service/s3.resolveEndpointV2Middleware`
  - ComputePayloadHash: `github.com/aws/aws-sdk-go-v2/aws/signer/v4.ComputePayloadSHA256`

      ```golang
      SetPayloadHash: middleware.WithStackValue(ctx, payloadHashKey{}, sha256.Sum(smithyReq.Stream()))
      ```

  - SigV4ContentSHA256Header: `github.com/aws/aws-sdk-go-v2/aws/signer/v4.ContentSHA256Header`

      ```golang
      req.Header["X-Amz-Content-Sha256"] = middleware.GetStackValue(ctx, payloadHashKey{}).(string)
      ```

  - Signing: `github.com/aws/aws-sdk-go-v2/service/s3.signRequestMiddleware`
- Deserialize
  - AddRawResponseToMetadata: `github.com/aws/aws-sdk-go-v2/aws/middleware.AddRawResponse`

    ```golang
    /* After */ metadata.Set(rawResponseKey{}, out.RawResponse)
    ```
  
  - ErrorCloseResponseBody: `github.com/aws/smithy-go/transport/http.errorCloseResponseBodyMiddleware`

    ```golang
    /* After */ if err != nil { io.Copy(io.Discard, resp.Body) }
    ```

  - CloseResponseBody: `github.com/aws/smithy-go/transport/http.closeResponseBody`

    ```golang
    /* After */ if err == nil { io.Copy(io.Discard, resp.Body); resp.Body.Close() }
    ```

  - ResponseErrorWrapper: `github.com/aws/aws-sdk-go-v2/service/internal/s3shared.errorWrapper`

    ```golang
    /* After */ if err != nil { return NewError(reqID, hostID, resp.Metadata)}
    ```

  - S3MetadataRetriever: `github.com/aws/aws-sdk-go-v2/service/internal/s3shared.metadataRetriever`

    ```golang
    /* After */
    resp.Metadata.SetRequestId(r.Header["X-Amz-Request-Id"])
    resp.Metadata.SetRequestId2(r.Header["X-Amz-Id2"])
    ```

  - OperationDeserializer: `github.com/aws/aws-sdk-go-v2/service/s3.awsRestxml_deserializeOpCreateBucket`

    ```golang
    /* After */ Deserialize request into the Output
    ```

  - RequestResponseLogger: `github.com/aws/smithy-go/transport/http.RequestResponseLogger`

    ```golang
    /* Before */ Log request
    /* After */ Log response
    ```
