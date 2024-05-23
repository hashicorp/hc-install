// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package install_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

func TestInstaller_Ensure_installable(t *testing.T) {
	testutil.EndToEndTest(t)

	// most of this logic is already tested within individual packages
	// so this is just a simple E2E test to ensure the public API
	// also works and continues working

	i := install.NewInstaller()
	i.SetLogger(testutil.TestLogger())
	ctx := context.Background()
	_, err := i.Ensure(ctx, []src.Source{
		&releases.LatestVersion{
			Product: product.Terraform,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Remove(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstaller_Ensure_findable(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")

	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dirPath)

	// most of this logic is already tested within individual packages
	// so this is just a simple E2E test to ensure the public API
	// also works and continues working

	i := install.NewInstaller()
	i.SetLogger(testutil.TestLogger())
	ctx := context.Background()
	_, err = i.Ensure(ctx, []src.Source{
		&fs.AnyVersion{
			Product: &product.Product{
				BinaryName: func() string {
					return fileName
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstaller_Install(t *testing.T) {
	testutil.EndToEndTest(t)

	// most of this logic is already tested within individual packages
	// so this is just a simple E2E test to ensure the public API
	// also works and continues working

	i := install.NewInstaller()
	i.SetLogger(testutil.TestLogger())
	ctx := context.Background()
	_, err := i.Install(ctx, []src.Installable{
		&releases.LatestVersion{
			Product: product.Terraform,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Remove(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstaller_Install_enterprise(t *testing.T) {
	testutil.EndToEndTest(t)

	// most of this logic is already tested within individual packages
	// so this is just a simple E2E test to ensure the public API
	// also works and continues working

	tmpBinaryDir := t.TempDir()
	tmpLicenseDir := t.TempDir()

	i := install.NewInstaller()
	i.SetLogger(testutil.TestLogger())
	ctx := context.Background()
	_, err := i.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product:    product.Vault,
			Version:    version.Must(version.NewVersion("1.9.8")),
			InstallDir: tmpBinaryDir,
			LicenseDir: tmpLicenseDir,
			Enterprise: &releases.EnterpriseOptions{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure the binary was installed
	binName := "vault"
	if runtime.GOOS == "windows" {
		binName = "vault.exe"
	}
	if _, err = os.Stat(filepath.Join(tmpBinaryDir, binName)); err != nil {
		t.Fatal(err)
	}
	// Ensure the enterprise license files were installed
	if _, err = os.Stat(filepath.Join(tmpLicenseDir, "EULA.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(tmpLicenseDir, "TermsOfEvaluation.txt")); err != nil {
		t.Fatal(err)
	}

	err = i.Remove(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
