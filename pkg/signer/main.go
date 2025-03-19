package signer

import (
	"time"

	"github.com/valyala/fasthttp"
)

type Signer interface {
	Sign(req *fasthttp.Request, region string, now time.Time) error
}
