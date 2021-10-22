package fs

import (
	"context"
	"os"
	"testing"

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
		Product: product.Product{
			BinaryName: func() string { return fileName },
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err := av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
