package client

import (
	"context"

	chain_of_responsibility "github.com/s3hobby/client/pkg/design-patterns/chain-of-responsibility"
)

type Client struct {
	options Options
}

func New(options *Options, optFns ...func(*Options)) (*Client, error) {
	c := &Client{}

	if options != nil {
		c.options = *options
	}

	for _, fn := range optFns {
		fn(&c.options)
	}

	c.options.setDefaults()
	if err := c.options.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func PerformCall[
	Input HttpRequestMarshaler,
	OutputPtr interface {
		HttpRequestUnmarshaler
		*OutputBase
	},
	OutputBase any,
](ctx context.Context, c *Client, input Input, optFns ...func(*Options)) (OutputPtr, error) {
	in := &handlerInput[Input]{
		Options:   c.options.With(optFns...),
		CallInput: input,
	}

	in.InitHTTP()
	defer in.ReleaseHTTP()

	chain := chain_of_responsibility.NewChain(
		&httpRequesterHandler[Input, OutputBase, OutputPtr]{},
		&errorMiddleware[Input, OutputPtr]{},
		&configValidationMiddleware[Input, OutputPtr]{},
		&requiredInputMiddleware[Input, OutputPtr]{},
		&userAgentMiddleware[Input, OutputPtr]{},
		&resolveEndpointMiddleware[Input, OutputPtr]{},
		&transportMiddleware[Input, OutputPtr]{},
		&signerMiddleware[Input, OutputPtr]{},
		// &serverSideErrorMiddleware[Input, OutputPtr]{},
	)

	out, err := chain.Handle(ctx, in)
	if err != nil {
		return nil, err
	}
	defer out.ReleaseHTTP()

	return out.CallOutput, nil
}
