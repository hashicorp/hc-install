// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package releases

// APIHTTPAuth holds optional credentials for HTTP requests to a custom
// releases mirror (ApiBaseURL). When BearerToken is non-empty, it is sent as
// Authorization: Bearer <token>. When Username is non-empty (and BearerToken
// is empty), HTTP basic authentication is used.
type APIHTTPAuth struct {
	Username    string
	Password    string
	BearerToken string
}
