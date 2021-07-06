package checkpoint

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Installable    = &LatestVersion{}
	_ src.Removable      = &LatestVersion{}
	_ src.LoggerSettable = &LatestVersion{}
)

func TestLatestVersion(t *testing.T) {
	testutil.EndToEndTest(t)

	lv := &LatestVersion{
		Product: product.Terraform,
	}
	lv.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	execPath, err := lv.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lv.Remove(ctx) })

	v, err := product.Terraform.GetVersion(ctx, execPath)
	if err != nil {
		t.Fatal(err)
	}

	latestConstraint, err := version.NewConstraint(">= 1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !latestConstraint.Check(v.Core()) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			latestConstraint, v)
	}
}
