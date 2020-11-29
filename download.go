package hcinstall

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	goGetter "github.com/hashicorp/go-getter"
	"golang.org/x/crypto/openpgp"
)

const releasesURL = "https://releases.hashicorp.com"

func downloadWithVerification(ctx context.Context, product string, productVersion string, installDir string, appendUserAgent string) (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	baseURL := releasesURL + "/" + product

	// setup: ensure we have a place to put our downloaded binary
	tfDir, err := ensureInstallDir(installDir)
	if err != nil {
		return "", err
	}

	httpGetter := &goGetter.HttpGetter{
		Netrc:  true,
		Client: newHTTPClient(appendUserAgent),
	}
	client := goGetter.Client{
		Ctx: ctx,
		Getters: map[string]goGetter.Getter{
			"https": httpGetter,
		},
	}
	client.Mode = goGetter.ClientModeAny

	// firstly, download and verify the signature of the checksum file

	sumsTmpDir, err := ioutil.TempDir("", "hcinstall")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(sumsTmpDir)

	sumsFilename := product + "_" + productVersion + "_SHA256SUMS"
	sumsSigFilename := sumsFilename + ".sig"

	sumsURL := fmt.Sprintf("%s/%s/%s", baseURL, productVersion, sumsFilename)
	sumsSigURL := fmt.Sprintf("%s/%s/%s", baseURL, productVersion, sumsSigFilename)

	client.Src = sumsURL
	client.Dst = sumsTmpDir
	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("error fetching checksums: %s", err)
	}

	client.Src = sumsSigURL
	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("error fetching checksums signature: %s", err)
	}

	sumsPath := filepath.Join(sumsTmpDir, sumsFilename)
	sumsSigPath := filepath.Join(sumsTmpDir, sumsSigFilename)

	err = verifySumsSignature(sumsPath, sumsSigPath)
	if err != nil {
		return "", err
	}

	// secondly, download the binary itself, verifying the checksum
	url := hcURL(product, productVersion, osName, archName)
	client.Src = url
	client.Dst = tfDir
	client.Mode = goGetter.ClientModeDir
	err = client.Get()
	if err != nil {
		return "", err
	}

	return filepath.Join(tfDir, product), nil
}

// verifySumsSignature downloads SHA256SUMS and SHA256SUMS.sig and verifies
// the signature using the HashiCorp public key.
func verifySumsSignature(sumsPath, sumsSigPath string) error {
	el, err := openpgp.ReadArmoredKeyRing(strings.NewReader(hashicorpPublicKey))
	if err != nil {
		return err
	}
	data, err := os.Open(sumsPath)
	if err != nil {
		return err
	}
	sig, err := os.Open(sumsSigPath)
	if err != nil {
		return err
	}
	_, err = openpgp.CheckDetachedSignature(el, data, sig)

	return err
}

func hcURL(product, productVersion, osName, archName string) string {
	sumsFilename := product + "_" + productVersion + "_SHA256SUMS"
	sumsURL := fmt.Sprintf("%s/%s/%s/%s", releasesURL, product, productVersion, sumsFilename)
	return fmt.Sprintf(
		"%s/%s/%s/%s_%s_%s_%s.zip?checksum=file:%s",
		releasesURL, product, productVersion, product, productVersion, osName, archName, sumsURL,
	)
}
