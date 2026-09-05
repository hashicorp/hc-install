// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package install_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	isrc "github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

// mockInstallable is a removable installable source for unit tests.
type mockInstallable struct {
	path       string
	installErr error
	removeErr  error

	installs atomic.Int32
	removes  atomic.Int32
}

func (*mockInstallable) IsSourceImpl() isrc.InstallSrcSigil {
	return isrc.InstallSrcSigil{}
}

func (m *mockInstallable) Install(ctx context.Context) (string, error) {
	m.installs.Add(1)
	return m.path, m.installErr
}

func (m *mockInstallable) Remove(ctx context.Context) error {
	m.removes.Add(1)
	return m.removeErr
}

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

func TestInstaller_Install_repeatedTracksRemovableSources(t *testing.T) {
	i := install.NewInstaller()
	ctx := context.Background()
	first := &mockInstallable{path: "/tmp/first"}
	second := &mockInstallable{path: "/tmp/second"}

	if _, err := i.Install(ctx, []src.Installable{first}); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Install(ctx, []src.Installable{second}); err != nil {
		t.Fatal(err)
	}
	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}

	if got := first.removes.Load(); got != 1 {
		t.Errorf("first source Remove calls = %d, want 1", got)
	}
	if got := second.removes.Load(); got != 1 {
		t.Errorf("second source Remove calls = %d, want 1", got)
	}
}

func TestInstaller_EnsureThenInstall_tracksBothRemovableSources(t *testing.T) {
	i := install.NewInstaller()
	ctx := context.Background()
	ensured := &mockInstallable{path: "/tmp/ensured"}
	installed := &mockInstallable{path: "/tmp/installed"}

	if _, err := i.Ensure(ctx, []src.Source{ensured}); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Install(ctx, []src.Installable{installed}); err != nil {
		t.Fatal(err)
	}
	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}

	if got := ensured.removes.Load(); got != 1 {
		t.Errorf("Ensure source Remove calls = %d, want 1", got)
	}
	if got := installed.removes.Load(); got != 1 {
		t.Errorf("Install source Remove calls = %d, want 1", got)
	}
}

func TestInstaller_Remove_clearsTrackedSources(t *testing.T) {
	i := install.NewInstaller()
	ctx := context.Background()
	srcA := &mockInstallable{path: "/tmp/a"}

	if _, err := i.Install(ctx, []src.Installable{srcA}); err != nil {
		t.Fatal(err)
	}
	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if got := srcA.removes.Load(); got != 1 {
		t.Errorf("Remove calls = %d, want 1 (tracked sources should be cleared)", got)
	}
}

func TestInstaller_EnsureInstallRemove_concurrent(t *testing.T) {
	i := install.NewInstaller()
	ctx := context.Background()
	ensureSrc := &mockInstallable{path: "/tmp/ensure"}
	installSrc := &mockInstallable{path: "/tmp/install"}

	const n = 32
	var wg sync.WaitGroup
	wg.Add(n * 3)
	for k := 0; k < n; k++ {
		go func() {
			defer wg.Done()
			_, _ = i.Ensure(ctx, []src.Source{ensureSrc})
		}()
		go func() {
			defer wg.Done()
			_, _ = i.Install(ctx, []src.Installable{installSrc})
		}()
		go func() {
			defer wg.Done()
			_ = i.Remove(ctx)
		}()
	}
	wg.Wait()

	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestInstaller_Install_concurrentTracksAllRemovableSources(t *testing.T) {
	i := install.NewInstaller()
	ctx := context.Background()

	const n = 32
	sources := make([]*mockInstallable, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for k := 0; k < n; k++ {
		sources[k] = &mockInstallable{path: "/tmp/mock"}
		go func(m *mockInstallable) {
			defer wg.Done()
			if _, err := i.Install(ctx, []src.Installable{m}); err != nil {
				t.Errorf("Install: %s", err)
			}
		}(sources[k])
	}
	wg.Wait()

	if err := i.Remove(ctx); err != nil {
		t.Fatal(err)
	}

	for k, m := range sources {
		if got := m.removes.Load(); got != 1 {
			t.Errorf("source %d Remove calls = %d, want 1", k, got)
		}
	}
}
