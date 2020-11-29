package hcinstall

import (
	"errors"

	"github.com/hashicorp/go-version"
)

type VersionConstraints struct {
	latest bool

	// UNIMPLEMENTED: see NewVersionConstraints
	constraints version.Constraints

	// DEPRECATED: to be removed when constraints is implemented, as
	// exact versions can be represented as version.Constraints.
	exact *version.Version
}

// NewVersionConstraints constructs a new version constraints, erroring if
// invalid. Constraints are parsed from strings such as ">= 1.0" using
// hashicorp/go-version. If the special string "latest" is supplied, the version
// is constrained to the latest version Checkpoint reports as available, which
// is determined at runtime during Install.
// TODO KEM: There is currently no way to find all versions of a product
// satisfying a constraint string such as ">=0.13.5". Add a new endpoint to
// Checkpoint that returns all available versions.
func NewVersionConstraints(constraints string) (*VersionConstraints, error) {
	if constraints == "latest" {
		return &VersionConstraints{
			latest: true,
		}, nil
	}

	exactVersion, err := version.NewSemver(constraints)
	if err != nil {
		return nil, errors.New("Error parsing version constraint %s: %w.\nPlease supply an exact semver version.")
	}

	return &VersionConstraints{
		exact: exactVersion,
	}, nil
}
