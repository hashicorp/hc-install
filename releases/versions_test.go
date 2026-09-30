// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/httpclient"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

func TestVersions_List_ApiBaseURL(t *testing.T) {
	mockApiRoot := filepath.Join("testdata", "mock_api_tf_0_14_with_prereleases")
	srv := testutil.NewTestServer(t, mockApiRoot)

	cons, err := version.NewConstraint(">= 0.14.0")
	if err != nil {
		t.Fatal(err)
	}

	versions := &Versions{
		Product:     product.Terraform,
		Constraints: cons,
		ApiBaseURL:  srv.URL,
	}

	sources, err := versions.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	ev := sources[0].(*ExactVersion)
	if ev.Version.String() != "0.14.11" {
		t.Fatalf("expected version 0.14.11, got %q", ev.Version.String())
	}
	if ev.ApiBaseURL != srv.URL {
		t.Fatalf("expected ExactVersion ApiBaseURL %q, got %q", srv.URL, ev.ApiBaseURL)
	}
	if ev.HTTPClient != nil {
		t.Fatal("expected ExactVersion HTTPClient to be left nil (default)")
	}
}

type bearerTokenRoundTripper struct {
	token string
	inner http.RoundTripper
}

func (rt *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+rt.token)
	return rt.inner.RoundTrip(req)
}

func TestVersions_List_HTTPClient(t *testing.T) {
	const token = "super-secret"

	mockApiRoot := filepath.Join("testdata", "mock_api_tf_0_14_with_prereleases")
	fileServer := http.FileServer(http.Dir(mockApiRoot))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	cons, err := version.NewConstraint("= 0.14.11")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("without auth", func(t *testing.T) {
		versions := &Versions{
			Product:     product.Terraform,
			Constraints: cons,
			ApiBaseURL:  srv.URL,
		}
		_, err := versions.List(ctx)
		if err == nil {
			t.Fatal("expected error listing versions from authenticated mirror without credentials")
		}
	})

	t.Run("with auth", func(t *testing.T) {
		client := httpclient.New(httpclient.WithLogger(testutil.TestLogger()))
		client.Transport = &bearerTokenRoundTripper{token: token, inner: client.Transport}

		versions := &Versions{
			Product:     product.Terraform,
			Constraints: cons,
			ApiBaseURL:  srv.URL,
			HTTPClient:  client,
			Install: InstallationOptions{
				ArmoredPublicKey: getTestPubKey(t),
			},
		}

		sources, err := versions.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(sources) != 1 {
			t.Fatalf("expected 1 source, got %d", len(sources))
		}

		ev := sources[0].(*ExactVersion)
		if ev.HTTPClient != client {
			t.Fatal("expected ExactVersion HTTPClient to be passed through from Versions")
		}

		// installation must reuse the same client to reach the mirror
		ev.SetLogger(testutil.TestLogger())
		execPath, err := ev.Install(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { ev.Remove(ctx) })

		v, err := product.Terraform.GetVersion(ctx, execPath)
		if err != nil {
			t.Fatal(err)
		}
		if !ev.Version.Equal(v) {
			t.Fatalf("versions don't match (expected: %s, installed: %s)", ev.Version, v)
		}
	})
}

func TestVersions_List(t *testing.T) {
	testutil.EndToEndTest(t)

	cons, err := version.NewConstraint(">= 1.0.0, < 1.0.10")
	if err != nil {
		t.Fatal(err)
	}

	versions := &Versions{
		Product:     product.Terraform,
		Constraints: cons,
	}

	ctx := context.Background()
	sources, err := versions.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	expectedVersions := []string{
		"1.0.0",
		"1.0.1",
		"1.0.2",
		"1.0.3",
		"1.0.4",
		"1.0.5",
		"1.0.6",
		"1.0.7",
		"1.0.8",
		"1.0.9",
	}
	if diff := cmp.Diff(expectedVersions, sourcesToRawVersions(sources)); diff != "" {
		t.Fatalf("unexpected versions: %s", diff)
	}
}

func TestVersions_List_enterprise(t *testing.T) {
	testutil.EndToEndTest(t)

	cons, err := version.NewConstraint(">= 1.9.0, < 1.9.9")
	if err != nil {
		t.Fatal(err)
	}

	versions := &Versions{
		Product:     product.Vault,
		Constraints: cons,
		Install: InstallationOptions{
			LicenseDir: "/some/path",
		},
		Enterprise: &EnterpriseOptions{
			Meta: "hsm",
		},
	}

	ctx := context.Background()
	sources, err := versions.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	expectedVersions := []string{
		"1.9.0+ent.hsm",
		"1.9.1+ent.hsm",
		"1.9.2+ent.hsm",
		"1.9.3+ent.hsm",
		"1.9.4+ent.hsm",
		"1.9.5+ent.hsm",
		"1.9.6+ent.hsm",
		"1.9.7+ent.hsm",
		"1.9.8+ent.hsm",
	}
	if diff := cmp.Diff(expectedVersions, sourcesToRawVersions(sources)); diff != "" {
		t.Fatalf("unexpected versions: %s", diff)
	}

	for _, source := range sources {
		if *source.(*ExactVersion).Enterprise != *versions.Enterprise {
			t.Fatalf("unexpected Enterprise data: %v", source.(*ExactVersion).Enterprise)
		}

		if source.(*ExactVersion).Enterprise == versions.Enterprise {
			t.Fatalf("the Enterprise data should be copied, not referenced")
		}
	}
}

func sourcesToRawVersions(srcs []src.Source) []string {
	rawVersions := make([]string, len(srcs))

	for idx, src := range srcs {
		source := src.(*ExactVersion)
		rawVersions[idx] = source.Version.String()
	}

	return rawVersions
}
