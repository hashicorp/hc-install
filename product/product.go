package product

import "github.com/hashicorp/go-version"

type Product struct {
	BinaryName string
	GetVersion func(execPath string) (*version.Version, error)
	RepoURL    string
}
