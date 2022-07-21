package chttp

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

const (
	// MethodGet HTTP method
	MethodGet = "GET"

	// MethodPost HTTP method
	MethodPost = "POST"

	// MethodPut HTTP method
	MethodPut = "PUT"

	// MethodDelete HTTP method
	MethodDelete = "DELETE"

	// MethodPatch HTTP method
	MethodPatch = "PATCH"

	// MethodHead HTTP method
	MethodHead = "HEAD"

	// MethodOptions HTTP method
	MethodOptions = "OPTIONS"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(*Client, *Request) error
)

type Client struct {
	BaseURL    string
	QueryParam url.Values
	Header     http.Header

	setContentLength bool
	trace            bool
	httpClient       *http.Client
	scheme           string
	closeConnection  bool
	beforeRequest    []RequestMiddleware
}

func CreateClient(hc *http.Client) *Client {
	if hc.Transport == nil {
		hc.Transport = createTransport(nil)
	}

	c := &Client{ // not setting lang default values
		QueryParam: url.Values{},
		Header:     http.Header{},
		httpClient: hc,
	}

	// default before request middlewares
	c.beforeRequest = []RequestMiddleware{
		createHTTPRequest,
	}
	return c
}

func createTransport(localAddr net.Addr) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

// R method creates a new request instance, its used for Get, Post, Put, Delete, Patch, Head, Options, etc.
func (c *Client) R() *Request {
	r := &Request{
		QueryParam: url.Values{},
		Header:     http.Header{},

		client: c,
	}
	return r
}

// NewRequest is an alias for method `R()`. Creates a new request instance, its used for
// Get, Post, Put, Delete, Patch, Head, Options, etc.
func (c *Client) NewRequest() *Request {
	return c.R()
}

// Executes method executes the given `Request` object and returns response
// error.
func (c *Client) execute(req *Request) (*Response, error) {
	// Apply Request middleware
	var err error

	// resty middlewares
	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, wrapNoRetryErr(err)
		}
	}

	if hostHeader := req.Header.Get("Host"); hostHeader != "" {
		req.RawRequest.Host = hostHeader
	}

	req.Time = time.Now()
	resp, err := c.httpClient.Do(req.RawRequest)
	response := &Response{
		Request:     req,
		RawResponse: resp,
	}
	if err != nil {
		response.setReceivedAt()
		return response, err
	}
	defer resp.Body.Close()
	response.body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		response.setReceivedAt()
		return response, err
	}
	response.setReceivedAt()
	return response, wrapNoRetryErr(err)
}
