package client

import "github.com/valyala/fasthttp"

func setQuery(args *fasthttp.Args, name string, value *string) {
	if value == nil {
		return
	}

	args.Set(name, *value)
}

func setHeader(requestHeader *fasthttp.RequestHeader, key string, value *string) {
	if value == nil {
		return
	}

	requestHeader.Set(key, *value)
}

func extractHeader(responseHeader *fasthttp.ResponseHeader, key string) *string {
	value := responseHeader.Peek(key)
	if value == nil {
		return nil
	}

	str := string(value)
	return &str
}
