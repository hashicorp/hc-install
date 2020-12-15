package hcinstall

import (
	"context"
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

func TestTerraform(t *testing.T) {
	v, err := NewVersionConstraints(">0.13.4, <0.13.6", true)
	if err != nil {
		t.Fatal(err)
	}

	client := &Client{
		Product:            products.Terraform,
		Getters:            []Getter{Releases()},
		VersionConstraints: v,
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
