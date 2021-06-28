package releases

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-version"
	rjson "github.com/hashicorp/hc-install/internal/releasesjson"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

// Versions allows listing all versions of a product
// which match Constraints
type Versions struct {
	Product     product.Product
	Constraints version.Constraints

	ListTimeout              time.Duration
	InstallTimeout           time.Duration
	InstallDir               string
	SkipChecksumVerification bool
}

func (v *Versions) List(ctx context.Context) ([]src.Source, error) {
	if v.Product.Name == "" {
		return nil, fmt.Errorf("unknown product name")
	}

	timeout := defaultListTimeout
	if v.ListTimeout > 0 {
		timeout = v.ListTimeout
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	r := rjson.NewReleases()
	pvs, err := r.ListProductVersions(ctx, v.Product.Name)
	if err != nil {
		return nil, err
	}

	installables := make([]src.Source, 0)
	for _, pv := range pvs {
		installableVersion, err := version.NewVersion(pv.Version)
		if err != nil {
			continue
		}

		if !v.Constraints.Check(installableVersion) {
			// skip version which doesn't match constraint
			continue
		}

		ev := &ExactVersion{
			Product:    v.Product,
			Version:    installableVersion,
			InstallDir: v.InstallDir,
			Timeout:    v.InstallTimeout,

			SkipChecksumVerification: v.SkipChecksumVerification,
		}

		installables = append(installables, ev)
	}

	return installables, nil
}
