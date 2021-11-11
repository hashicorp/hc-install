package install

import (
	"context"
	"testing"

	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

func TestInstaller_Ensure(t *testing.T) {
	testutil.EndToEndTest(t)

	// most of this logic is already tested within individual packages
	// so this is just a simple E2E test to ensure the public API
	// also works and continues working

	i := NewInstaller()
	_, err := i.Ensure(context.Background(), []src.Source{
		&releases.LatestVersion{
			Product: product.Terraform,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
