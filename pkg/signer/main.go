package signer

import (
	"github.com/valyala/fasthttp"
)

type Signer interface {
	Sign(req *fasthttp.Request, region string) error
}
