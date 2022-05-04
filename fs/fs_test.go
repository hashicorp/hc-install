package fs

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/magodo/hc-install/internal/testutil"
	"github.com/magodo/hc-install/product"
	"github.com/magodo/hc-install/src"
)

var (
	_ src.Findable       = &AnyVersion{}
	_ src.LoggerSettable = &AnyVersion{}

	_ src.Findable       = &ExactVersion{}
	_ src.LoggerSettable = &ExactVersion{}
)

func TestExactVersion(t *testing.T) {
	t.Skip("TODO")
	testutil.EndToEndTest(t)

	// TODO: mock out command execution?

	t.Setenv("PATH", "")

	ev := &ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.0")),
	}
	ev.SetLogger(testutil.TestLogger())
	_, err := ev.Find(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
