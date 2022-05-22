package version

import "runtime/debug"

// ModuleVersion returns the current version of the github.com/hashicorp/hc-install Go module.
func ModuleVersion() string {
	version := "0.0.0-devel"

	bi, ok := debug.ReadBuildInfo()
	if ok {
		version = bi.Main.Version
	}

	return version
}
