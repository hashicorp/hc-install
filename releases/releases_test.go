// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Installable = &ExactVersion{}
	_ src.Removable   = &ExactVersion{}

	_ src.Installable = &LatestVersion{}
	_ src.Removable   = &LatestVersion{}
)

func TestLatestVersion(t *testing.T) {
	testutil.EndToEndTest(t)

	lv := &LatestVersion{
		Product: product.Terraform,
	}
	lv.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	installDetails, err := lv.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}
	v := installDetails.Version
	t.Cleanup(func() { lv.Remove(ctx) })

	latestConstraint, err := version.NewConstraint(">= 1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !latestConstraint.Check(v) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			latestConstraint, v)
	}
	if installDetails.Product != lv.Product.Name {
		t.Fatalf("expected product name %q, got %q", lv.Product.Name, installDetails.Product)
	}
}

func TestLatestVersion_basic(t *testing.T) {
	mockApiRoot := filepath.Join("testdata", "mock_api_tf_0_14_with_prereleases")
	lv := &LatestVersion{
		Product:          product.Terraform,
		ArmoredPublicKey: getTestPubKey(t),
		ApiBaseURL:       testutil.NewTestServer(t, mockApiRoot).URL,
	}
	lv.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	installDetails, err := lv.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}
	v := installDetails.Version
	t.Cleanup(func() { lv.Remove(ctx) })

	expectedVersion, err := version.NewVersion("0.14.11")
	if err != nil {
		t.Fatal(err)
	}
	if !expectedVersion.Equal(v) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			expectedVersion, v)
	}
	if installDetails.Product != lv.Product.Name {
		t.Fatalf("expected product name %q, got %q", lv.Product.Name, installDetails.Product)
	}
}

func TestLatestVersion_prereleases(t *testing.T) {
	mockApiRoot := filepath.Join("testdata", "mock_api_tf_0_14_with_prereleases")

	lv := &LatestVersion{
		Product:            product.Terraform,
		IncludePrereleases: true,
		ArmoredPublicKey:   getTestPubKey(t),
		ApiBaseURL:         testutil.NewTestServer(t, mockApiRoot).URL,
	}
	lv.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	installDetails, err := lv.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}
	v := installDetails.Version

	t.Cleanup(func() { lv.Remove(ctx) })

	expectedVersion, err := version.NewVersion("0.15.0-rc2")
	if err != nil {
		t.Fatal(err)
	}
	if !expectedVersion.Equal(v) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			expectedVersion, v)
	}
}

func TestExactVersion(t *testing.T) {
	testutil.EndToEndTest(t)

	versionToInstall := version.Must(version.NewVersion("1.8.2"))
	ev := &ExactVersion{
		Product: product.Terraform,
		Version: versionToInstall,
	}
	ev.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	installDetails, err := ev.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}

	execPath := installDetails.ExecutablePath

	licensePath := filepath.Join(filepath.Dir(execPath), "LICENSE.txt")
	t.Cleanup(func() {
		ev.Remove(ctx)
		// check if license was deleted
		if _, err := os.Stat(licensePath); !os.IsNotExist(err) {
			t.Fatalf("license file not deleted at %q: %s", licensePath, err)
		}
	})

	t.Logf("exec path of installed: %s", execPath)

	v, err := product.Terraform.GetVersion(ctx, execPath)
	if err != nil {
		t.Fatal(err)
	}

	if !versionToInstall.Equal(installDetails.Version) {
		t.Fatalf("the version of the installed binary %s does not match the returned version %s",
			installDetails.Version, versionToInstall)
	}

	if !versionToInstall.Equal(v) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			versionToInstall, v)
	}

	// check if license was copied
	if _, err := os.Stat(licensePath); err != nil {
		t.Fatalf("expected license file not found at %q: %s", licensePath, err)
	}
}

func BenchmarkExactVersion(b *testing.B) {
	mockApiRoot := filepath.Join("testdata", "mock_api_tf_0_14_with_prereleases")

	for i := 0; i < b.N; i++ {
		installDir, err := os.MkdirTemp("", fmt.Sprintf("%s_%d", "terraform", i))
		if err != nil {
			b.Fatal(err)
		}

		ev := &ExactVersion{
			Product:          product.Terraform,
			Version:          version.Must(version.NewVersion("0.14.11")),
			ArmoredPublicKey: getTestPubKey(b),
			ApiBaseURL:       testutil.NewTestServer(b, mockApiRoot).URL,
			InstallDir:       installDir,
		}
		ev.SetLogger(testutil.TestLogger())

		ctx := context.Background()
		_, err = ev.Install(ctx)
		if err != nil {
			b.Fatal(err)
		}
		b.Cleanup(func() { ev.Remove(ctx) })
	}
}

func getTestPubKey(t testing.TB) string {
	f, err := os.Open(filepath.Join("testdata", "2FCA0A85.pub"))
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
