// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	rjson "github.com/hashicorp/hc-install/internal/releasesjson"
	"github.com/hashicorp/hc-install/product"
)

func TestLatestVersionValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		lv          LatestVersion
		expectedErr error
	}{
		"Product-incorrect-binary-name": {
			lv: LatestVersion{
				Product: product.Product{
					BinaryName: func() string { return "invalid!" },
					Name:       product.Terraform.Name,
				},
			},
			expectedErr: fmt.Errorf("invalid binary name: \"invalid!\""),
		},
		"Product-incorrect-name": {
			lv: LatestVersion{
				Product: product.Product{
					BinaryName: product.Terraform.BinaryName,
					Name:       "invalid!",
				},
			},
			expectedErr: fmt.Errorf("invalid product name: \"invalid!\""),
		},
		"Product-valid": {
			lv: LatestVersion{
				Product: product.Terraform,
			},
		},
		"Enterprise-missing-license-dir": {
			lv: LatestVersion{
				Product:    product.Vault,
				Enterprise: &EnterpriseOptions{},
			},
			expectedErr: fmt.Errorf("LicenseDir must be provided when requesting enterprise versions"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.lv.Validate()

			if err == nil && testCase.expectedErr != nil {
				t.Fatalf("expected error: %s, got no error", testCase.expectedErr)
			}

			if err != nil && testCase.expectedErr == nil {
				t.Fatalf("expected no error, got error: %s", err)
			}

			if err != nil && testCase.expectedErr != nil && err.Error() != testCase.expectedErr.Error() {
				t.Fatalf("expected error: %s, got error: %s", testCase.expectedErr, err)
			}
		})
	}
}

func TestLatestVersion_FindLatestMatchingVersion(t *testing.T) {
	t.Parallel()

	possibleVersions := rjson.ProductVersionsMap{
		"1.14.0": &rjson.ProductVersion{
			Version: version.Must(version.NewVersion("1.14.0")),
		},
		"1.14.1": &rjson.ProductVersion{
			Version: version.Must(version.NewVersion("1.14.1")),
		},
		"1.15.2": &rjson.ProductVersion{
			Version: version.Must(version.NewVersion("1.15.2")),
		},
		"1.14.1+ent": &rjson.ProductVersion{
			Version: version.Must(version.NewVersion("1.14.1+ent")),
		},
		"1.14.1+ent.fips1402": &rjson.ProductVersion{
			Version: version.Must(version.NewVersion("1.14.1+ent.fips1402")),
		},
	}
	constraints, _ := version.NewConstraint("~> 1.14.0")

	testCases := map[string]struct {
		lv              LatestVersion
		expectedVersion string
	}{
		"oss1": {
			lv: LatestVersion{
				Product: product.Vault,
			},
			expectedVersion: "1.15.2",
		},
		"oss2": {
			lv: LatestVersion{
				Product:     product.Vault,
				Constraints: constraints,
			},
			expectedVersion: "1.14.1",
		},
		"enterprise": {
			lv: LatestVersion{
				Product:    product.Vault,
				Enterprise: &EnterpriseOptions{},
			},
			expectedVersion: "1.14.1+ent",
		},
		"enterprise-fips1402": {
			lv: LatestVersion{
				Product: product.Vault,
				Enterprise: &EnterpriseOptions{
					Meta: "fips1402",
				},
			},
			expectedVersion: "1.14.1+ent.fips1402",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			latest, _ := testCase.lv.findLatestMatchingVersion(possibleVersions, testCase.lv.Constraints)

			if latest.Version.Original() != testCase.expectedVersion {
				t.Fatalf("expected version %s, got %s", testCase.expectedVersion, latest.Version.Original())
			}
		})
	}
}
