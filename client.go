package client

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
