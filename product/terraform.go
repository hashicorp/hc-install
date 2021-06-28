package product

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	simpleVersionRe = `v?(?P<version>[0-9]+(?:\.[0-9]+)*(?:-[A-Za-z0-9\.]+)?)`

	terraformVersionOutputRe = regexp.MustCompile(`Terraform ` + simpleVersionRe)
)

var Terraform = Product{
	BinaryName: "terraform",
	GetVersion: func(path string) (*version.Version, error) {
		cmd := exec.Command(path, "version")

		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		stdout := strings.TrimSpace(string(out))

		submatches := terraformVersionOutputRe.FindStringSubmatch(stdout)
		if len(submatches) != 2 {
			return nil, fmt.Errorf("unexpected number of version matches %d for %s", len(submatches), stdout)
		}
		v, err := version.NewVersion(submatches[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse version %q: %w", submatches[1], err)
		}

		return v, err

	},
	RepoURL: "https://github.com/hashicorp/terraform.git",
}
