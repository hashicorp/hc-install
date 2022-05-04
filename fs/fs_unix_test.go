//go:build !windows
// +build !windows

package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/magodo/hc-install/errors"
	"github.com/magodo/hc-install/internal/testutil"
	"github.com/magodo/hc-install/product"
)

func TestAnyVersion_notExecutable(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	t.Setenv("PATH", dirPath)

	av := &AnyVersion{
		Product: &product.Product{
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

	dirPath, fileName := testutil.CreateTempFile(t, "")
	t.Setenv("PATH", dirPath)

	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		Product: &product.Product{
			BinaryName: func() string { return fileName },
		},
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnyVersion_constraint(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	t.Setenv("PATH", dirPath)

	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		Product: &product.Product{
			BinaryName: func() string { return fileName },
			GetVersion: func(ctx context.Context, execPath string) (*version.Version, error) {
				return version.NewVersion("1.2.0")
			},
		},
		Constraints: version.MustConstraints(version.NewConstraint("~> 1.0")),
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnyVersion_constraintNotMet(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	t.Setenv("PATH", dirPath)

	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		Product: &product.Product{
			BinaryName: func() string { return fileName },
			GetVersion: func(ctx context.Context, execPath string) (*version.Version, error) {
				return version.NewVersion("2.0.0")
			},
		},
		Constraints: version.MustConstraints(version.NewConstraint("~> 1.0")),
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err == nil {
		t.Fatal("expected error for non-executable file")
	}

	if !errors.IsErrorSkippable(err) {
		t.Fatalf("expected a skippable error, got: %#v", err)
	}
}

func TestAnyVersion_exactBinPath(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		ExactBinPath: fullPath,
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnyVersion_exactBinPath_notExecutable(t *testing.T) {
	testutil.EndToEndTest(t)

	dirPath, fileName := testutil.CreateTempFile(t, "")
	fullPath := filepath.Join(dirPath, fileName)
	err := os.Chmod(fullPath, 0600)
	if err != nil {
		t.Fatal(err)
	}

	av := &AnyVersion{
		ExactBinPath: fullPath,
	}
	av.SetLogger(testutil.TestLogger())
	_, err = av.Find(context.Background())
	if err == nil {
		t.Fatal("expected error for non-executable file")
	}

	if !errors.IsErrorSkippable(err) {
		t.Fatalf("expected a skippable error, got: %#v", err)
	}
}
