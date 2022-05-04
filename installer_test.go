package install

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/hc-install/fs"
	"github.com/magodo/hc-install/internal/testutil"
	"github.com/magodo/hc-install/product"
	"github.com/magodo/hc-install/releases"
	"github.com/magodo/hc-install/src"
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
