package hcinstall

import (
	"context"
)

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

	if g.c.VersionConstraints.latest {
		// use checkpoint to find the latest version and use that as productVersion
	} else {
		productVersion = g.c.VersionConstraints.exact.String()
	}

	p, err := downloadWithVerification(ctx, g.c.Product.Name, productVersion, g.c.InstallDir, g.appendUserAgent)
	if err != nil {
		return "", err
	}

	return p, nil
}
