package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/hc-install/errors"
	"github.com/magodo/hc-install/internal/testutil"
	"github.com/magodo/hc-install/product"
)

func TestAnyVersion_executable(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	t.Setenv("path", dirPath)

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

	dirPath, fileName := testutil.CreateTempFile(t, "")
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

	dirPath, fileName := testutil.CreateTempFile(t, "")
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
