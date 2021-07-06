package build

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/testutil"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

var (
	_ src.Buildable      = &GitRevision{}
	_ src.Removable      = &GitRevision{}
	_ src.LoggerSettable = &GitRevision{}
)

func TestGitRevision_terraform(t *testing.T) {
	testutil.EndToEndTest(t)

	gr := &GitRevision{Product: product.Terraform}
	gr.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	execPath, err := gr.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { gr.Remove(ctx) })

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

func TestGitRevision_consul(t *testing.T) {
	testutil.EndToEndTest(t)

	gr := &GitRevision{Product: product.Consul}
	gr.SetLogger(testutil.TestLogger())

	ctx := context.Background()

	execPath, err := gr.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { gr.Remove(ctx) })

	v, err := product.Consul.GetVersion(ctx, execPath)
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
