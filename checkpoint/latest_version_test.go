// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package checkpoint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Installable    = &LatestVersion{}
	_ src.Removable      = &LatestVersion{}
	_ src.LoggerSettable = &LatestVersion{}
)

func TestLatestVersion(t *testing.T) {
	testutil.EndToEndTest(t)

	lv := &LatestVersion{
		Product: product.Terraform,
	}
	lv.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	installDetails, err := lv.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}
	execPath := installDetails.ExecutablePath
	v := installDetails.Version

	licensePath := filepath.Join(filepath.Dir(execPath), "LICENSE.txt")
	t.Cleanup(func() {
		lv.Remove(ctx)
		// check if license was deleted
		if _, err := os.Stat(licensePath); !os.IsNotExist(err) {
			t.Fatalf("license file not deleted at %q: %s", licensePath, err)
		}
	})

	latestConstraint, err := version.NewConstraint(">= 1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !latestConstraint.Check(v.Core()) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			latestConstraint, v)
	}

	// check if license was copied
	if _, err := os.Stat(licensePath); err != nil {
		t.Fatalf("expected license file not found at %q: %s", licensePath, err)
	}
}

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
