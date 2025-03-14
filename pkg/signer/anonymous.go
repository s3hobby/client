package signer

import "github.com/valyala/fasthttp"

func NewAnonymousSigner() Signer {
	return &anonymousSigner{}
}

type anonymousSigner struct {
}

func (*anonymousSigner) Sign(*fasthttp.Request, string) error {
	return nil
}
