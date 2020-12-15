package hcinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcinstall/products"
)

const releasesURL = "https://releases.hashicorp.com"

func Releases() *ReleasesGetter {
	return &ReleasesGetter{}
}

func (g *ReleasesGetter) AppendUserAgent(ua string) {
	g.appendUserAgent = ua
}

type ReleasesGetter struct {
	getter
	appendUserAgent string
}

func (g *ReleasesGetter) Get(ctx context.Context) (string, error) {
	var productVersion string
	var err error

	if g.c.VersionConstraints.latest {
		productVersion, err = getLatestVersion(g.c.Product, g.c.VersionConstraints.forceCheckpoint)
		if err != nil {
			return "", err
		}
	} else if g.c.VersionConstraints.exact != nil {
		productVersion = g.c.VersionConstraints.exact.String()
	} else {
		productVersion, err = getLatestVersionMatchingConstraints(g.c.Product, g.c.VersionConstraints.constraints)
		if err != nil {
			return "", err
		}
	}

	p, err := downloadWithVerification(ctx, g.c.Product.Name, productVersion, g.c.InstallDir, g.appendUserAgent)
	if err != nil {
		return "", err
	}

	return p, nil
}

func getLatestVersion(product products.Product, forceCheckpoint bool) (string, error) {
	resp, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: product.Name,
		Force:   forceCheckpoint,
	})
	if err != nil {
		return "", err
	}

	if resp.CurrentVersion == "" {
		return "", fmt.Errorf("could not determine latest version using checkpoint: CHECKPOINT_DISABLE may be set")
	}

	return resp.CurrentVersion, nil
}

// Product is a top-level product like "Consul" or "Nomad". A Product may have
// one or more versions.
type releasedProduct struct {
	Name     string                             `json:"name"`
	Versions map[string]*releasedProductVersion `json:"versions"`
}

// ProductVersion is a wrapper around a particular product version like
// "consul 0.5.1". A ProductVersion may have one or more builds.
type releasedProductVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// SHASUMS    string          `json:"shasums,omitempty"`
	// SHASUMSSig string          `json:"shasums_signature,omitempty"`
	// Builds     []*ProductBuild `json:"builds"`
}

func getLatestVersionMatchingConstraints(product products.Product, constraints version.Constraints) (string, error) {
	allProductVersions := releasedProduct{}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	r, err := httpClient.Get(releasesURL + "/" + product.Name + "/index.json")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&allProductVersions)
	if err != nil {
		return "", err
	}

	// allProductVersions is an unsorted list of all available versions:
	// we must therefore visit each one to determine the maximum version
	// satisfying the constraints

	zeroVersion, err := version.NewVersion("0.0.0")
	if err != nil {
		return "", fmt.Errorf("Unexpected error parsing initial value of maxVersion: this is a bug in hcinstall")
	}

	maxVersion := zeroVersion

	for v := range allProductVersions.Versions {
		vers, err := version.NewVersion(v)
		if err != nil {
			// hc-releases runs all versions through version.NewVersion,
			// so something is seriously wrong if we can't parse here
			return "", fmt.Errorf("Error parsing releases version %s: %s", v, err)
		}

		if constraints.Check(vers) {
			if vers.GreaterThan(maxVersion) {
				maxVersion = vers
			}
		}
	}

	if maxVersion == zeroVersion {
		return "", fmt.Errorf("No version of %s found satisfying version constraints %s", product.Name, constraints)
	}

	return maxVersion.String(), nil
}
