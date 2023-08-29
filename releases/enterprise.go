// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"fmt"
)

type EnterpriseOptions struct {
	LicenseDir string // required
	Meta       string // optional; may be "hsm", "fips1402", "hsm.fips1402", etc.
}

func (eo *EnterpriseOptions) requiredMetadata() string {
	metadata := "ent"
	if eo.Meta != "" {
		metadata += "." + eo.Meta
	}
	return metadata
}

func (eo *EnterpriseOptions) validate() error {
	if eo.LicenseDir == "" {
		return fmt.Errorf("LicenseDir must be provided when requesting enterprise versions")
	}
	return nil
}
