// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package httpclient

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/hc-install/version"
)

// NewHTTPClient provides a pre-configured http.Client
// e.g. with relevant User-Agent header.
//
// If transport is non-nil, it is called with hc-install's own transport
// (which already handles retries and sets the User-Agent header) and the
// http.RoundTripper it returns becomes the client's Transport. This allows
// callers to wrap outgoing requests, e.g. to inject authentication headers
// required by a private mirror.
func NewHTTPClient(logger *log.Logger, transport func(http.RoundTripper) http.RoundTripper) *http.Client {
	rc := retryablehttp.NewClient()
	rc.Logger = logger
	client := rc.StandardClient()
	client.Transport = &userAgentRoundTripper{
		userAgent: fmt.Sprintf("hc-install/%s", version.Version()),
		inner:     client.Transport,
	}
	if transport != nil {
		client.Transport = transport(client.Transport)
	}
	return client
}

type userAgentRoundTripper struct {
	inner     http.RoundTripper
	userAgent string
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", rt.userAgent)
	}
	return rt.inner.RoundTrip(req)
}
