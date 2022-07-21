package chttp

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	URL        string
	Method     string
	QueryParam url.Values
	Header     http.Header
	Time       time.Time
	Body       interface{}
	Result     interface{}
	Error      interface{}
	RawRequest *http.Request

	setContentLength bool
	trace            bool
	ctx              context.Context
	values           map[string]interface{}
	client           *Client
	bodyBuf          *bytes.Buffer
	clientTrace      *clientTrace
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTTP request tracing
//_______________________________________________________________________

// EnableTrace method enables trace for the current request
// using `httptrace.ClientTrace` and provides insights.
//
// 		client := resty.New()
//
// 		resp, err := client.R().EnableTrace().Get("https://httpbin.org/get")
// 		fmt.Println("Error:", err)
// 		fmt.Println("Trace Info:", resp.Request.TraceInfo())
//
// See `Client.EnableTrace` available too to get trace info for all requests.
//
// Since v2.0.0
func (r *Request) EnableTrace() *Request {
	r.trace = true
	return r
}

// TraceInfo method returns the trace info for the request.
// If either the Client or Request EnableTrace function has not been called
// prior to the request being made, an empty TraceInfo object will be returned.
//
// Since v2.0.0
func (r *Request) TraceInfo() TraceInfo {
	ct := r.clientTrace

	if ct == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		DNSLookup:      ct.dnsDone.Sub(ct.dnsStart),
		TLSHandshake:   ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		ServerTime:     ct.gotFirstResponseByte.Sub(ct.gotConn),
		IsConnReused:   ct.gotConnInfo.Reused,
		IsConnWasIdle:  ct.gotConnInfo.WasIdle,
		ConnIdleTime:   ct.gotConnInfo.IdleTime,
		RequestAttempt: 0,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if ct.gotConnInfo.Reused {
		ti.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		ti.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}

// Context method returns the Context if its already set in request
// otherwise it creates new one using `context.Background()`.
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

func (r *Request) selectAddr(addrs []*net.SRV, path string, attempt int) string {
	if addrs == nil {
		return path
	}

	idx := attempt % len(addrs)
	domain := strings.TrimRight(addrs[idx].Target, ".")
	path = strings.TrimLeft(path, "/")

	return fmt.Sprintf("%s://%s:%d/%s", r.client.scheme, domain, addrs[idx].Port, path)
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTTP verb method starts here
//_______________________________________________________________________

// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (r *Request) Get(url string) (*Response, error) {
	return r.Execute(MethodGet, url)
}

// Execute method performs the HTTP request with given HTTP method and URL
// for current `Request`.
// 		resp, err := client.R().Execute(resty.GET, "http://httpbin.org/get")
func (r *Request) Execute(method, url string) (*Response, error) {
	var addrs []*net.SRV
	var resp *Response
	var err error

	r.Method = method
	r.URL = r.selectAddr(addrs, url, 0)

	resp, err = r.client.execute(r)
	return resp, unwrapNoRetryErr(err)
}
