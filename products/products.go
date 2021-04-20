package products

import (
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

	// RepoURL is the URL for the product's git repo.
	RepoURL string
}
