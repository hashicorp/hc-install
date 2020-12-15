package hcinstall

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcinstall/products"
)

func TestGetLatestVersionMatchingConstraints(t *testing.T) {
	for i, c := range []struct {
		constraints           string
		product               products.Product
		expectedLatestVersion string
	}{
		{
			"0.12.26", products.Terraform, "0.12.26",
		},
		{
			"<0.11.4", products.Terraform, "0.11.3",
		},
		{
			">0.13.0, <0.13.2", products.Terraform, "0.13.1",
		},
		{
			">0.12.0-alpha4, <0.12.0-rc2", products.Terraform, "0.12.0-rc1",
		},
	} {
		t.Run(fmt.Sprintf("%d %s", i, c.expectedLatestVersion), func(t *testing.T) {
			cs, err := version.NewConstraint(c.constraints)
			if err != nil {
				t.Fatal(err)
			}

			v, err := getLatestVersionMatchingConstraints(c.product, cs)
			if err != nil {
				t.Fatal(err)
			}

			if v != c.expectedLatestVersion {
				t.Fatalf("expected %s, got %s", c.expectedLatestVersion, v)
			}
		})
	}
}

func TestGetLatestVersionMatchingConstraints_no_available_version(t *testing.T) {
	cs, err := version.NewConstraint(">0.999.999, <1.0")
	if err != nil {
		t.Fatal(err)
	}

	v, err := getLatestVersionMatchingConstraints(products.Terraform, cs)
	if err == nil {
		t.Fatalf("Expected getLatestVersionMatchingConstraints to error, but it did not (returned version: %s)", v)
	}
}
