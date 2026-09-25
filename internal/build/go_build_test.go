// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestReadGoModVersionUsesToolchainDirective(t *testing.T) {
	repoDir := t.TempDir()
	goMod := `module example.com/product

go 1.24.0
toolchain go1.26.8
`
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok := readGoModVersion(repoDir)
	if !ok {
		t.Fatal("readGoModVersion returned no version")
	}
	want := version.Must(version.NewVersion("1.26.8"))
	if !got.Equal(want) {
		t.Fatalf("readGoModVersion() = %s, want %s", got, want)
	}
}

func TestReadGoModVersionFallsBackToGoDirective(t *testing.T) {
	repoDir := t.TempDir()
	goMod := `module example.com/product

go 1.24.0
future-directive value
`
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	got, ok := readGoModVersion(repoDir)
	if !ok {
		t.Fatal("readGoModVersion returned no version")
	}
	want := version.Must(version.NewVersion("1.24.0"))
	if !got.Equal(want) {
		t.Fatalf("readGoModVersion() = %s, want %s", got, want)
	}
}
