package client

import (
	"encoding/xml"
	"fmt"

	"github.com/valyala/fasthttp"
)

type ClientSideError struct {
	OperationName string
	Err           error
}

func (e *ClientSideError) Unwrap() error {
	return e.Err
}

func (e *ClientSideError) Error() string {
	return fmt.Sprintf("client-side error occurred when calling the %s operation: %v", e.OperationName, e.Err)
}

type ServerSideError struct {
	OperationName string `xml:"-"`
	Code          string `xml:"Code"`
	Message       string `xml:"Message"`
	RequestID     string `xml:"RequestId"`
	HostID        string `xml:"HostId"`

	Response *fasthttp.Response `xml:"-"`
}

func (e *ServerSideError) Error() string {
	var code string
	if e.Code != "" {
		code = fmt.Sprintf("(Code:%s) ", e.Code)
	}

	var requestID string
	if e.RequestID != "" {
		requestID = fmt.Sprintf("(RequestID:%s) ", e.RequestID)
	}

	var hostID string
	if e.HostID != "" {
		hostID = fmt.Sprintf("(HostID:%s) ", e.HostID)
	}

	return fmt.Sprintf(
		"server-side error occurred %s%s%swhen calling the %s operation: %s",
		code,
		requestID,
		hostID,
		e.OperationName,
		e.Message,
	)
}

func NewServerSideError(operationName string, resp *fasthttp.Response) (*ServerSideError, error) {
	ret := new(ServerSideError)

	if err := xml.Unmarshal(resp.Body(), ret); err != nil {
		return nil, fmt.Errorf("ServerSideError: xml error response deserializing error: %w", err)
	}

	ret.Response = new(fasthttp.Response)
	resp.CopyTo(ret.Response)

	ret.OperationName = operationName

	return ret, nil
}
