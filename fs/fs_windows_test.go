package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hc-install/errors"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
)

func TestAnyVersion_executable(t *testing.T) {
	testutil.EndToEndTest(t)

	originalPath := os.Getenv("path")
	os.Setenv("path", "")
	t.Cleanup(func() {
		os.Setenv("path", originalPath)
	})

	dirPath, fileName := createTempFile(t, "")
	os.Setenv("path", dirPath)

	av := &AnyVersion{
		Product: &product.Product{
			BinaryName: func() string { return fileName },
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err := av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnyVersion_exactBinPath(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := createTempFile(t, "")
	fullPath := filepath.Join(dirPath, fileName)

	av := &AnyVersion{
		ExactBinPath: fullPath,
	}
	av.SetLogger(testutil.TestLogger())
	_, err := av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnyVersion_exactBinPath_notFound(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := createTempFile(t, "")
	fullPath := filepath.Join(dirPath, fileName)

	err := os.Remove(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		ExactBinPath: fullPath,
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !errors.IsErrorSkippable(err) {
		t.Fatalf("expected a skippable error, got: %#v", err)
	}
}
