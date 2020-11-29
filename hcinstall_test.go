package hcinstall

import (
	"context"
	// "io/ioutil"
	// "os"
	"os/exec"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	// tmpDir, err := ioutil.TempDir("", "hcinstall-test")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer os.RemoveAll(tmpDir)

	tfPath, err := Install(context.Background(), "", ProductTerraform, "0.12.26")
	if err != nil {
		t.Fatal(err)
	}

	// run "terraform version" to check we've downloaded a terraform 0.12.26 binary
	cmd := exec.Command(tfPath, "version")

	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform v0.12.26"
	actual := string(out)
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("ran terraform version, expected %s, but got %s", expected, actual)
	}
}
