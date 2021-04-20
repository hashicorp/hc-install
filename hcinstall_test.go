package hcinstall

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/hcinstall/products"
)

func TestInstall(t *testing.T) {
	tfPath, err := Install(context.Background(), "", products.Terraform, "0.12.26", true)
	if err != nil {
		t.Fatal(err)
	}

	// run "terraform version" to check we've downloaded a terraform 0.12.26 binary
	cmd := exec.Command(tfPath, "version")

	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform v0.12.26"
	actual := string(out)
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("ran terraform version, expected %s, but got %s", expected, actual)
	}
}

func TestConsul_releases(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "hcinstall-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	v, err := NewVersionConstraints(">1.9.3, <1.9.5", true)
	if err != nil {
		t.Fatal(err)
	}

	client := &Client{
		Product:            products.Consul,
		Getters:            []Getter{Releases()},
		VersionConstraints: v,
		InstallDir:         tmpDir,
	}

	consulPath, err := client.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// run "terraform version" to check we've downloaded a terraform 0.12.26 binary
	cmd := exec.Command(consulPath, "version")

	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	expected := "Consul v1.9.4"
	actual := string(out)
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("ran consul version, expected %s, but got %s", expected, actual)
	}
}

func TestTerraform_releases(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "hcinstall-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	v, err := NewVersionConstraints(">0.13.4, <0.13.6", true)
	if err != nil {
		t.Fatal(err)
	}

	client := &Client{
		Product:            products.Terraform,
		Getters:            []Getter{Releases()},
		VersionConstraints: v,
		InstallDir:         tmpDir,
	}

	tfPath, err := client.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// run "terraform version" to check we've downloaded a terraform 0.12.26 binary
	cmd := exec.Command(tfPath, "version")

	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform v0.13.5"
	actual := string(out)
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("ran terraform version, expected %s, but got %s", expected, actual)
	}
}

func TestTerraform_gitref(t *testing.T) {
	for i, c := range []struct {
		gitRef          string
		expectedVersion string
	}{
		{"refs/heads/main", "Terraform v0.15.0-dev"},
		{"refs/tags/v0.12.29", "Terraform v0.12.29"},
		{"refs/pull/26921/head", "Terraform v0.14.0-dev"},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "hcinstall-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// v, err := NewVersionConstraints(">=0.1.0-a", true)
			// if err != nil {
			// 	t.Fatal(err)
			// }

			client := &Client{
				Product: products.Terraform,
				Getters: []Getter{GitRef(c.gitRef)},
				// VersionConstraints: v,
				InstallDir:          tmpDir,
				DisableVersionCheck: true,
			}

			tfPath, err := client.Install(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(tfPath, "version")

			out, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}

			actual := string(out)
			if !strings.HasPrefix(actual, c.expectedVersion) {
				t.Fatalf("ran terraform version, expected %s, but got %s", c.expectedVersion, actual)
			}
		})
	}
}

func TestTerraform_gitcommit(t *testing.T) {
	for i, c := range []struct {
		gitCommit       string
		expectedVersion string
	}{
		// using CHANGELOG commits, since these are unlikely to be removed
		{"45b795b3fd02d5177666218c3703e26252eeb745", "Terraform v0.15.0-dev"},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "hcinstall-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// v, err := NewVersionConstraints(">=0.1.0-a", true)
			// if err != nil {
			// 	t.Fatal(err)
			// }

			client := &Client{
				Product: products.Terraform,
				Getters: []Getter{GitCommit(c.gitCommit)},
				// VersionConstraints: v,
				InstallDir:          tmpDir,
				DisableVersionCheck: true,
			}

			tfPath, err := client.Install(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(tfPath, "version")

			out, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}

			// expected := "Terraform v0.15.0-dev"
			actual := string(out)
			if !strings.HasPrefix(actual, c.expectedVersion) {
				t.Fatalf("ran terraform version, expected %s, but got %s", c.expectedVersion, actual)
			}
		})
	}
}
