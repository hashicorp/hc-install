package hcinstall

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

// Product is a HashiCorp product downloadable via hcinstall.
type Product struct {
	// Name is the name of the binary to be installed, also used as a
	// friendly name in log messages.
	Name string

	// GetVersion tries to determine the version of the executable at the
	// supplied path.
	GetVersion func(string) (*version.Version, error)
}

var ProductTerraform = Product{
	Name: "terraform",
	GetVersion: func(path string) (*version.Version, error) {
		cmd := exec.Command(path, "version")

		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		return parseTerraformVersionOutput(string(out))
	},
}

var (
	simpleVersionRe = `v?(?P<version>[0-9]+(?:\.[0-9]+)*(?:-[A-Za-z0-9\.]+)?)`

	terraformVersionOutputRe = regexp.MustCompile(`Terraform ` + simpleVersionRe)
)

func parseTerraformVersionOutput(stdout string) (*version.Version, error) {
	stdout = strings.TrimSpace(stdout)

	submatches := terraformVersionOutputRe.FindStringSubmatch(stdout)
	if len(submatches) != 2 {
		return nil, fmt.Errorf("unexpected number of version matches %d for %s", len(submatches), stdout)
	}
	v, err := version.NewVersion(submatches[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse version %q: %w", submatches[1], err)
	}

	return v, err
}
