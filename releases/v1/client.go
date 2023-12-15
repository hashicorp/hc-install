// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"io"
	"log"
	"time"

	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-releases/generated/client"
	"github.com/hashicorp/go-releases/generated/client/operations"
)

var (
	defaultInstallTimeout = 30 * time.Second
	defaultListTimeout    = 10 * time.Second
	discardLogger         = log.New(io.Discard, "", 0)
)

func newClient() operations.ClientService {
	ct := runtimeclient.New(
		client.DefaultHost,
		client.DefaultBasePath,
		client.DefaultSchemes)
	ct.Consumers["application/vnd+hashicorp.releases-api.v1+json"] = runtime.JSONConsumer()
	ct.Producers["application/vnd+hashicorp.releases-api.v1+json"] = runtime.JSONProducer()
	return client.New(ct, strfmt.Default).Operations
}
