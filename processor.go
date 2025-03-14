package client

import (
	"context"
	"errors"
	"fmt"

	chain_of_responsibility "github.com/s3hobby/client/pkg/design-patterns/chain-of-responsibility"

	"github.com/valyala/fasthttp"
)

type Handler[Input any, Output any] = chain_of_responsibility.Handler[*handlerInput[Input], *handlerOutput[Output]]

type httpRequestMarshaler interface {
	marshalHTTP(req *fasthttp.Request) error
}

type httpRequestUnmarshaler interface {
	unmarshalHTTP(resp *fasthttp.Response) error
}

type requiredBucketInterface interface {
	bucket() string
}

type requiredBucketKeyInterface interface {
	requiredBucketInterface
	key() string
}

type handlerInput[Input any] struct {
	// OperationName     string
	Options           *Options
	SuccessStatusCode int

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

func handleCall[
	Input httpRequestMarshaler,
	OutputPtr interface {
		httpRequestUnmarshaler
		*OutputBase
	},
	OutputBase any,
](ctx context.Context, input *handlerInput[Input]) (*handlerOutput[OutputPtr], error) {
	chain := chain_of_responsibility.NewChain(
		&httpRequesterHandler[Input, OutputBase, OutputPtr]{},
		&errorMiddleware[Input, OutputPtr]{},
		&configValidationMiddleware[Input, OutputPtr]{},
		&requiredInputMiddleware[Input, OutputPtr]{},
		&userAgentMiddleware[Input, OutputPtr]{},
		&resolveEndpointMiddleware[Input, OutputPtr]{},
		&transportMiddleware[Input, OutputPtr]{},
		&signerMiddleware[Input, OutputPtr]{},
		&serverSideErrorMiddleware[Input, OutputPtr]{},
	)

	output, err := chain.Handle(ctx, input)
	if err != nil {
		return nil, err
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

	if v, ok := interface{}(input.CallInput).(requiredBucketKeyInterface); ok {
		params.Bucket = v.bucket()
		params.Key = v.key()
	} else if v, ok := interface{}(input.CallInput).(requiredBucketInterface); ok {
		params.Bucket = v.bucket()
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
	if err := input.Options.Signer.Sign(input.ServerRequest, input.Options.SiginingRegion); err != nil {
		return nil, fmt.Errorf("cannot sign the request: %v", err)
	}

	return next.Handle(ctx, input)
}

type transportMiddleware[Input httpRequestMarshaler, Output httpRequestUnmarshaler] struct{}

func (*transportMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if err := input.CallInput.marshalHTTP(input.ServerRequest); err != nil {
		return nil, fmt.Errorf("HTTP marshaling error: %v", err)
	}

	output, err := next.Handle(ctx, input)
	if err != nil {
		return nil, err
	}

	if err = output.CallOutput.unmarshalHTTP(output.ServerResponse); err != nil {
		return nil, fmt.Errorf("HTTP unmarshalling error: %v", err)
	}

	return output, nil
}

type requiredInputMiddleware[Input any, Output any] struct{}

func (*requiredInputMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	if v, ok := interface{}(input.CallInput).(requiredBucketKeyInterface); ok {
		if v.bucket() == "" {
			return nil, errors.New("bucket is mandatory")
		}

		if v.key() == "" {
			return nil, errors.New("object key is mandatory")
		}
	} else if v, ok := interface{}(input.CallInput).(requiredBucketInterface); ok {
		if v.bucket() == "" {
			return nil, errors.New("bucket is mandatory")
		}
	}

	return next.Handle(ctx, input)
}

type serverSideErrorMiddleware[Input any, Output any] struct{}

func (*serverSideErrorMiddleware[Input, Output]) Middleware(ctx context.Context, input *handlerInput[Input], next Handler[Input, Output]) (*handlerOutput[Output], error) {
	output, err := next.Handle(ctx, input)
	if err != nil {
		return nil, err
	}

	statusCode := output.ServerResponse.StatusCode()
	if statusCode == input.SuccessStatusCode {
		return output, nil
	}

	sse, err := NewServerSideError(output.ServerResponse)
	if err != nil {
		return nil, err
	}

	return nil, sse
}
