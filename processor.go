package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	chain_of_responsibility "github.com/s3hobby/client/pkg/design-patterns/chain-of-responsibility"

	"github.com/valyala/fasthttp"
)

type Handler[Input any, Output any] = chain_of_responsibility.Handler[*handlerInput[Input], *handlerOutput[Output]]

type HttpRequestMarshaler interface {
	MarshalHTTP(req *fasthttp.Request) error
}

type HttpRequestUnmarshaler interface {
	UnmarshalHTTP(resp *fasthttp.Response) error
}

type RequiredBucketInterface interface {
	GetBucket() string
}

var _ RequiredBucketInterface = (RequiredBucketKeyInterface)(nil)

type RequiredBucketKeyInterface interface {
	RequiredBucketInterface
	GetKey() string
}

type handlerInput[Input any] struct {
	Options       *Options
	CallInput     Input
	ServerRequest *fasthttp.Request
}

func (input *handlerInput[Input]) InitHTTP() {
	input.ServerRequest = fasthttp.AcquireRequest()
}

func (input *handlerInput[Input]) ReleaseHTTP() {
	if input.ServerRequest == nil {
		return
	}

	fasthttp.ReleaseRequest(input.ServerRequest)
	input.ServerRequest = nil
}

type handlerOutput[Output any] struct {
	CallOutput     Output
	ServerResponse *fasthttp.Response
}

func (output *handlerOutput[Output]) ReleaseHTTP() {
	if output.ServerResponse == nil {
		return
	}

	fasthttp.ReleaseResponse(output.ServerResponse)
	output.ServerResponse = nil
}

type httpRequesterHandler[Input any, OutputBase any, OutputPtr *OutputBase] struct{}

func (*httpRequesterHandler[Input, OutputBase, OutputPtr]) Handle(ctx context.Context, input *handlerInput[Input]) (*handlerOutput[OutputPtr], error) {
	var callOutputBase OutputBase
	output := &handlerOutput[OutputPtr]{
		CallOutput:     &callOutputBase,
		ServerResponse: fasthttp.AcquireResponse(),
	}

	if err := input.Options.HTTPClient.Do(input.ServerRequest, output.ServerResponse); err != nil {
		output.ReleaseHTTP()
		return nil, fmt.Errorf("HTTP request error: %v", err)
	}

	return output, nil
}

type errorMiddleware[Input any, Output any] struct{}

func (*errorMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	output, err := next.Handle(ctx, input)
	if err == nil {
		return output, nil
	}

	var serverSideError *ServerSideError
	var clientSideError *ClientSideError

	if !errors.As(err, &serverSideError) && !errors.As(err, &clientSideError) {
		err = &ClientSideError{Err: err}
	}

	return output, err
}

type configValidationMiddleware[Input any, Output any] struct{}

func (*configValidationMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if err := input.Options.validate(); err != nil {
		return nil, err
	}

	return next.Handle(ctx, input)
}

type userAgentMiddleware[Input any, Output any] struct{}

func (*userAgentMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if input.Options.UserAgent != nil && *input.Options.UserAgent != "" {
		input.ServerRequest.Header.SetUserAgent(*input.Options.UserAgent)
	}

	return next.Handle(ctx, input)
}

type resolveEndpointMiddleware[Input any, Output any] struct{}

func (*resolveEndpointMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	params := EndpointParameters{
		Host:         input.Options.EndpointHost,
		UseSSL:       input.Options.UseSSL,
		UsePathStyle: input.Options.UsePathStyle,
	}

	if v, ok := interface{}(input.CallInput).(RequiredBucketKeyInterface); ok {
		params.Bucket = v.GetBucket()
		params.Key = v.GetKey()
	} else if v, ok := interface{}(input.CallInput).(RequiredBucketInterface); ok {
		params.Bucket = v.GetBucket()
	}

	endpoint, err := input.Options.EndpointResolver.ResolveEndpoint(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve endpoint: %v", err)
	}

	input.ServerRequest.SetRequestURI(endpoint.URL)

	return next.Handle(ctx, input)
}

type signerMiddleware[Input any, Output any] struct{}

func (*signerMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if _, _, err := input.Options.Signer.Sign(input.ServerRequest, input.Options.Credentials, input.Options.SiginingRegion, time.Now()); err != nil {
		return nil, fmt.Errorf("cannot sign the request: %v", err)
	}

	return next.Handle(ctx, input)
}

type transportMiddleware[Input HttpRequestMarshaler, Output HttpRequestUnmarshaler] struct{}

func (*transportMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if err := input.CallInput.MarshalHTTP(input.ServerRequest); err != nil {
		return nil, fmt.Errorf("HTTP marshaling error: %v", err)
	}

	output, err := next.Handle(ctx, input)
	if err != nil {
		return nil, err
	}

	if err = output.CallOutput.UnmarshalHTTP(output.ServerResponse); err != nil {
		// Do not wrap error since an unexpected HTTP status code can make
		// UnmarshalHTTP to return a server-side error.
		return nil, err
	}

	return output, nil
}

type requiredInputMiddleware[Input any, Output any] struct{}

func (*requiredInputMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if v, ok := interface{}(input.CallInput).(RequiredBucketKeyInterface); ok {
		if v.GetBucket() == "" {
			return nil, errors.New("bucket is mandatory")
		}

		if v.GetKey() == "" {
			return nil, errors.New("object key is mandatory")
		}
	} else if v, ok := interface{}(input.CallInput).(RequiredBucketInterface); ok {
		if v.GetBucket() == "" {
			return nil, errors.New("bucket is mandatory")
		}
	}

	return next.Handle(ctx, input)
}
