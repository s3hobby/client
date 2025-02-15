package signer

import (
	"github.com/valyala/fasthttp"
)

type Signer interface {
	Sign(req *fasthttp.Request, region string) error
}

func NewAnonymousSigner() Signer {
	return &anonymousSigner{}
}

type anonymousSigner struct {
}

func (*anonymousSigner) Sign(*fasthttp.Request, string) error {
	return nil
}
