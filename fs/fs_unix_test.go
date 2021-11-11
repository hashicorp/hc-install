//go:build !windows
// +build !windows

package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
)

func TestAnyVersion_notExecutable(t *testing.T) {
	testutil.EndToEndTest(t)

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	t.Cleanup(func() {
		os.Setenv("PATH", originalPath)
	})

	dirPath, fileName := createTempFile(t, "")
	os.Setenv("PATH", dirPath)

	av := &AnyVersion{
		Product: product.Product{
			BinaryName: func() string { return fileName },
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err := av.Find(context.Background())
	if err == nil {
		t.Fatalf("expected %s not to be found in %s", fileName, dirPath)
	}
}

func TestAnyVersion_executable(t *testing.T) {
	testutil.EndToEndTest(t)

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	t.Cleanup(func() {
		os.Setenv("PATH", originalPath)
	})

	dirPath, fileName := createTempFile(t, "")
	os.Setenv("PATH", dirPath)

	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		Product: product.Product{
			BinaryName: func() string { return fileName },
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
