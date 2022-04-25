package install

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
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

	i := NewInstaller()
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

	i := NewInstaller()
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

	i := NewInstaller()
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

func TestInstaller_InstallAndRemove(t *testing.T) {
	ctx := context.Background()
	installDir := filepath.Join(t.TempDir(), "hcinstall")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(installDir); err != nil {
			t.Fatal(err)
		}
	})

	i := NewInstaller()
	v := version.Must(version.NewVersion("1.0.2"))

	_, err := i.Ensure(ctx, []src.Source{
		&fs.ExactVersion{
			Product: product.Terraform,
			Version: v,
		},
		&releases.ExactVersion{
			Product:    product.Terraform,
			Version:    v,
			InstallDir: installDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := i.Remove(ctx); err != nil {
			t.Fatal(err)
		}
	})
}
