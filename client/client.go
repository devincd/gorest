package client

import (
	"devincd.io/gorest/client/chttp"
	"net/http"
)

func NewWithHTTPClient(hc *http.Client) *chttp.Client {
	if hc == nil {
		hc = http.DefaultClient
	}
	return chttp.CreateClient(hc)
}
