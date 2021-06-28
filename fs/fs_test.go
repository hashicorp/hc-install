package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Findable       = &AnyVersion{}
	_ src.LoggerSettable = &AnyVersion{}

	_ src.Findable       = &ExactVersion{}
	_ src.LoggerSettable = &ExactVersion{}
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
			BinaryName: fileName,
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
			BinaryName: fileName,
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestExactVersion(t *testing.T) {
	t.Skip("TODO")
	testutil.EndToEndTest(t)

	// TODO: mock out command execution?

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	t.Cleanup(func() {
		os.Setenv("PATH", originalPath)
	})

	ev := &ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.0")),
	}
	ev.SetLogger(testutil.TestLogger())
	_, err := ev.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func createTempFile(t *testing.T, content string) (string, string) {
	tmpDir := t.TempDir()
	fileName := t.Name()

	filePath := filepath.Join(tmpDir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}

	return tmpDir, fileName
}
