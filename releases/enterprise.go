// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"fmt"
)

type EnterpriseOptions struct {
	Enterprise bool
	Meta       string // optional; may be "hsm", "fips1402", "hsm.fips1402", etc.
	LicenseDir string // required when Enterprise is true
}

func (eo *EnterpriseOptions) requiredMetadata() string {
	metadata := ""
	if eo.Enterprise {
		metadata += "ent"
	}
	if eo.Meta != "" {
		metadata += "." + eo.Meta
	}
	return metadata
}

func (eo *EnterpriseOptions) validate() error {
	if eo.Enterprise && eo.LicenseDir == "" {
		return fmt.Errorf("license dir must be provided when requesting enterprise versions")
	}
	return nil
}
