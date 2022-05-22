package version

import "runtime/debug"

// ModuleVersion returns the current version of the github.com/hashicorp/hc-install Go module.
func ModuleVersion() string {
	version := "0.0.0-devel"

	bi, ok := debug.ReadBuildInfo()
	if ok && bi.Main.Version != "" {
		version = bi.Main.Version
	}
	if ok && bi.Main.Sum != "" {
		version = bi.Main.Sum
	}

	return version
}
