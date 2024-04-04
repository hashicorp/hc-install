// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releasesjson

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
)

func TestListProductVersions_includesEnterpriseBuilds(t *testing.T) {
	testutil.EndToEndTest(t)

	r := NewReleases()
	r.SetLogger(testutil.TestLogger())

	ctx := context.Background()
	pVersions, err := r.ListProductVersions(ctx, "consul")
	if err != nil {
		t.Fatal(err)
	}

	testEntVersion := "1.9.8+ent"
	_, ok := pVersions[testEntVersion]
	if !ok {
		t.Fatalf("Failed to find expected Consul Enterprise version %q", testEntVersion)
	}
}

func TestGetProductVersion_includesEnterpriseBuild(t *testing.T) {
	testutil.EndToEndTest(t)

	r := NewReleases()
	r.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	testEntVersion := version.Must(version.NewVersion("1.9.8+ent"))

	version, err := r.GetProductVersion(ctx, "consul", testEntVersion)
	if err != nil {
		t.Fatalf("Unexpected error getting enterprise version %q",
			testEntVersion.String())
	}

	if version.Version.String() != testEntVersion.Original() {
		t.Fatalf("Expected version %q, got %q", testEntVersion.String(), version.Version.String())
	}
}
