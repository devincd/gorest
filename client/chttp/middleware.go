package chttp

import (
	"io"
	"net/http"
)

func createHTTPRequest(c *Client, r *Request) (err error) {
	if r.bodyBuf == nil {
		if reader, ok := r.Body.(io.Reader); ok {
			r.RawRequest, err = http.NewRequest(r.Method, r.URL, reader)
		} else if c.setContentLength || r.setContentLength {
			r.RawRequest, err = http.NewRequest(r.Method, r.URL, http.NoBody)
		} else {
			r.RawRequest, err = http.NewRequest(r.Method, r.URL, nil)
		}
	} else {
		r.RawRequest, err = http.NewRequest(r.Method, r.URL, r.bodyBuf)
	}

	if err != nil {
		return
	}

	// Assign close connection option
	r.RawRequest.Close = c.closeConnection

	// Add headers into http request
	r.RawRequest.Header = r.Header

	// Enable trace
	if c.trace || r.trace {
		r.clientTrace = &clientTrace{}
		r.ctx = r.clientTrace.createContext(r.Context())
	}

	// Use context if it was specified
	if r.ctx != nil {
		r.RawRequest = r.RawRequest.WithContext(r.ctx)
	}
	return
}
