package products

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

var consulVersionOutputRe = regexp.MustCompile(`Consul ` + simpleVersionRe)

var Consul = Product{
	Name: "consul",
	GetVersion: func(path string) (*version.Version, error) {
		cmd := exec.Command(path, "version")

		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		stdout := strings.TrimSpace(string(out))

		submatches := consulVersionOutputRe.FindStringSubmatch(stdout)
		if len(submatches) != 2 {
			return nil, fmt.Errorf("unexpected number of version matches %d for %s", len(submatches), stdout)
		}
		v, err := version.NewVersion(submatches[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse version %q: %w", submatches[1], err)
		}

		return v, err
	},
	RepoURL: "https://github.com/hashicorp/consul.git",
}
