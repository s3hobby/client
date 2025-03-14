package client

import (
	"errors"
	"fmt"
	"testing"

	"github.com/s3hobby/client/pkg/fasthttptesting"
	"github.com/s3hobby/client/pkg/signer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type noMandatoryInput struct {
	QueryValue  string
	HeaderValue string
}

func (input *noMandatoryInput) marshalHTTP(req *fasthttp.Request) error {
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("x-header-value", input.HeaderValue)
	req.URI().QueryArgs().Set("query-value", input.QueryValue)
	return nil
}

type noMandatoryOutput struct {
	OneOutput string
}

func (output *noMandatoryOutput) unmarshalHTTP(resp *fasthttp.Response) error {
	if resp.StatusCode() != fasthttp.StatusNoContent {
		return fmt.Errorf("noResourceOutput: bad HTTP status code: %d", resp.StatusCode())
	}

	if value := resp.Header.Peek("x-header-value"); value != nil {
		output.OneOutput = string(value)
	}

	return nil
}

type mandatoryBucketInput struct {
	noMandatoryInput
	Bucket string
}

func (input *mandatoryBucketInput) bucket() string {
	return input.Bucket
}

type mandatoryKeyInput struct {
	mandatoryBucketInput
	Key string
}

func (input *mandatoryKeyInput) key() string {
	return input.Key
}

func testHandleCall_ok[Input httpRequestMarshaler](t *testing.T, apiIn Input, expectedURI string) {
	expectedOut := &noMandatoryOutput{
		OneOutput: uuid.NewString(),
	}

	srv := fasthttptesting.NewInmemoryTester(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, expectedURI, ctx.URI().String())

		ctx.Response.Header.Set("x-header-value", expectedOut.OneOutput)
		ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
	})
	defer srv.Close()

	in := &handlerInput[Input]{
		Options: &Options{
			SiginingRegion:   "dev-1",
			EndpointHost:     "s3.dev-1.example.com",
			Signer:           signer.NewAnonymousSigner(),
			HTTPClient:       srv.Client(),
			EndpointResolver: DefaultEndpointResolver,
		},
		SuccessStatusCode: fasthttp.StatusNoContent,
		CallInput:         apiIn,
	}
	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[Input, *noMandatoryOutput](t.Context(), in)
	require.NoError(t, err)
	defer out.ReleaseHTTP()

	require.NotNil(t, out)
	require.Equal(t, expectedOut, out.CallOutput)
}

func testHandleCall_ko[Input httpRequestMarshaler](t *testing.T, apiIn Input, expectedError error) {
	in := &handlerInput[Input]{
		Options: &Options{
			SiginingRegion:   "dev-1",
			EndpointHost:     "s3.dev-1.example.com",
			Signer:           signer.NewAnonymousSigner(),
			HTTPClient:       DefaultHTTPClient,
			EndpointResolver: DefaultEndpointResolver,
		},
		SuccessStatusCode: fasthttp.StatusNoContent,
		CallInput:         apiIn,
	}
	in.InitHTTP()
	defer in.ReleaseHTTP()

	out, err := handleCall[Input, *noMandatoryOutput](t.Context(), in)
	require.Error(t, err)
	require.Nil(t, out)

	var actual *ClientSideError
	require.ErrorAs(t, err, &actual)

	expected := &ClientSideError{Err: expectedError}
	require.Equal(t, expected, actual)
}

func Test_handleCall(t *testing.T) {
	t.Run("no resouces", func(t *testing.T) {
		apiIn := &noMandatoryInput{
			QueryValue:  uuid.NewString(),
			HeaderValue: uuid.NewString(),
		}

		testHandleCall_ok(
			t,
			apiIn,
			"http://s3.dev-1.example.com/?query-value="+apiIn.QueryValue,
		)
	})

	t.Run("with bucket", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			apiIn := &mandatoryBucketInput{
				noMandatoryInput: noMandatoryInput{
					QueryValue:  uuid.NewString(),
					HeaderValue: uuid.NewString(),
				},
				Bucket: uuid.NewString(),
			}

			testHandleCall_ok(
				t,
				apiIn,
				"http://"+apiIn.Bucket+".s3.dev-1.example.com/?query-value="+apiIn.QueryValue,
			)
		})

		t.Run("missing bucket", func(t *testing.T) {
			apiIn := &mandatoryBucketInput{}
			testHandleCall_ko(t, apiIn, errors.New("bucket is mandatory"))
		})
	})

	t.Run("with key", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			apiIn := &mandatoryKeyInput{
				mandatoryBucketInput: mandatoryBucketInput{
					noMandatoryInput: noMandatoryInput{
						QueryValue:  uuid.NewString(),
						HeaderValue: uuid.NewString(),
					},
					Bucket: uuid.NewString(),
				},
				Key: uuid.NewString(),
			}

			testHandleCall_ok(
				t,
				apiIn,
				"http://"+apiIn.Bucket+".s3.dev-1.example.com/"+apiIn.Key+"?query-value="+apiIn.QueryValue,
			)
		})

		t.Run("missing bucket", func(t *testing.T) {
			apiIn := &mandatoryKeyInput{
				Key: uuid.NewString(),
			}

			testHandleCall_ko(t, apiIn, errors.New("bucket is mandatory"))
		})

		t.Run("missing key", func(t *testing.T) {
			apiIn := &mandatoryKeyInput{
				mandatoryBucketInput: mandatoryBucketInput{
					Bucket: uuid.NewString(),
				},
			}

			testHandleCall_ko(t, apiIn, errors.New("object key is mandatory"))
		})

		t.Run("missing all", func(t *testing.T) {
			apiIn := &mandatoryKeyInput{}
			testHandleCall_ko(t, apiIn, errors.New("bucket is mandatory"))
		})
	})
}
