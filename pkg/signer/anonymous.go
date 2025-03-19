package signer

import (
	"time"

	"github.com/valyala/fasthttp"
)

func NewAnonymousSigner() Signer {
	return &anonymousSigner{}
}

type anonymousSigner struct {
}

func (*anonymousSigner) Sign(*fasthttp.Request, string, time.Time) error {
	return nil
}
