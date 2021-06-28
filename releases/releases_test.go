package releases

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Installable = &ExactVersion{}
	_ src.Removable   = &ExactVersion{}

	_ src.Installable = &LatestVersion{}
	_ src.Removable   = &LatestVersion{}
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

func TestExactVersion(t *testing.T) {
	testutil.EndToEndTest(t)

	versionToInstall := version.Must(version.NewVersion("0.14.0"))
	ev := &ExactVersion{
		Product: product.Terraform,
		Version: versionToInstall,
	}
	ev.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	execPath, err := ev.Install(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { ev.Remove(ctx) })

	t.Logf("exec path of installed: %s", execPath)

	v, err := product.Terraform.GetVersion(ctx, execPath)
	if err != nil {
		t.Fatal(err)
	}

	if !versionToInstall.Equal(v) {
		t.Fatalf("versions don't match (expected: %s, installed: %s)",
			versionToInstall, v)
	}
}
