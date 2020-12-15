package hcinstall

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
)

type VersionConstraints struct {
	latest bool

	constraints version.Constraints

	exact *version.Version

	forceCheckpoint bool
}

// NewVersionConstraints constructs a new version constraints, erroring if
// invalid. Constraints are parsed from strings such as ">= 1.0" using
// hashicorp/go-version. If the special string "latest" is supplied, the version
// is constrained to the latest version Checkpoint reports as available, which
// is determined at runtime during Install.
// Multiple constraints such as ">=1.2, < 1.0" are supported. Please see the
// documentation for hashicorp/go-version for more details.
func NewVersionConstraints(constraints string, forceCheckpoint bool) (*VersionConstraints, error) {
	if constraints == "latest" {
		return &VersionConstraints{
			latest: true,
		}, nil
	}

	// we treat single exact version constraints as a special case
	// to save a network request in Get
	exactVersionRegexp := regexp.MustCompile(`^=?(` + version.SemverRegexpRaw + `)$`)
	matches := exactVersionRegexp.FindStringSubmatch(constraints)
	if matches != nil {
		v, err := version.NewSemver(matches[2])
		if err != nil {
			return nil, fmt.Errorf("Error parsing version %s: %s", constraints, err)
		}
		return &VersionConstraints{
			exact: v,
		}, nil
	}

	c, err := version.NewConstraint(constraints)
	if err != nil {
		return nil, fmt.Errorf("Error parsing version constraints %s: %s", constraints, err)
	}

	return &VersionConstraints{
		constraints: c,
	}, nil
}
