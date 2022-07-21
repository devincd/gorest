package chttp

import (
	"net/http"
	"strings"
	"time"
)

// Response struct holds response values of executed request.
type Response struct {
	Request     *Request
	RawResponse *http.Response

	body       []byte
	size       int64
	receivedAt time.Time
}

// Status method returns the HTTP status string for the executed request.
//	Example: 200 OK
func (r *Response) Status() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

// StatusCode method returns the HTTP status code for the executed request.
//	Example: 200
func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

// Proto method returns the HTTP response protocol used for the request.
func (r *Response) Proto() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

// Time method returns the time of HTTP response time that from request we sent and received a request.
//
// See `Response.ReceivedAt` to know when client received response and see `Response.Request.Time` to know
// when client sent a request.
func (r *Response) Time() time.Duration {
	if r.Request.clientTrace != nil {
		return r.Request.TraceInfo().TotalTime
	}
	return r.receivedAt.Sub(r.Request.Time)
}

// ReceivedAt method returns when response got received from server for the request.
func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

// String method returns the body of the server response as String.
func (r *Response) String() string {
	if r.body == nil {
		return ""
	}
	return strings.TrimSpace(string(r.body))
}

func (r *Response) setReceivedAt() {
	r.receivedAt = time.Now()
	if r.Request.clientTrace != nil {
		r.Request.clientTrace.endTime = r.receivedAt
	}
}
