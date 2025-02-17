package client

import (
	"github.com/s3hobby/client/pkg/signer"

	"github.com/go-playground/validator/v10"
)

type Options struct {
	UserAgent string

	UsePathStyle bool
	EndpointHost string `validate:"hostname|hostname_port"`
	UseSSL       bool

	// EndpointResolver default to [DefaultEndpointResolver].
	EndpointResolver EndpointResolver `validate:"required"`

	SiginingRegion string `validate:"required"`

	Signer signer.Signer `validate:"required"`

	// HTTPClient default to [DefaultHTTPClient].
	HTTPClient HTTPClient `validate:"required"`
}

// With return a new instance of [Options] with applied transformations.
func (opts *Options) With(optFns ...func(*Options)) *Options {
	var ret Options = *opts

	for _, fn := range optFns {
		fn(&ret)
	}

	return &ret
}

func (opts *Options) setDefaults() {
	if opts.EndpointResolver == nil {
		opts.EndpointResolver = &defaultEndpointResolver{}
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = DefaultHTTPClient
	}
}

func (opts *Options) validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(opts)
}
