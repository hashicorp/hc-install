package releasesjson

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/hashicorp/hc-install/internal/httpclient"
)

type Downloader struct {
	Logger         *log.Logger
	VerifyChecksum bool
}

func (d *Downloader) DownloadAndUnpack(ctx context.Context, pv *ProductVersion, dstDir string) error {
	if len(pv.Builds) == 0 {
		return fmt.Errorf("no builds found for %s %s", pv.Name, pv.Version)
	}

	pb, ok := pv.Builds.BuildForOsArch(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return fmt.Errorf("no build found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	var verifiedChecksum HashSum
	if d.VerifyChecksum {
		v := &ChecksumDownloader{
			ProductVersion: pv,
			Logger:         d.Logger,
		}
		verifiedChecksums, err := v.DownloadAndVerifyChecksums()
		if err != nil {
			return err
		}
		var ok bool
		verifiedChecksum, ok = verifiedChecksums[pb.Filename]
		if !ok {
			return fmt.Errorf("no checksum found for %q", pb.Filename)
		}
	}

	client := httpclient.NewHTTPClient()
	resp, err := client.Get(pb.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var pkgReader io.Reader
	pkgReader = resp.Body

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code (%d)", resp.StatusCode)
	}

	contentType := resp.Header.Get("content-type")
	if contentType != "application/zip" {
		return fmt.Errorf("unexpected content-type: %s (expected application/zip)",
			contentType)
	}

	if d.VerifyChecksum {
		d.Logger.Printf("calculating checksum of %q", pb.Filename)
		// provide extra reader to calculate & compare checksum
		var buf bytes.Buffer
		r := io.TeeReader(resp.Body, &buf)
		pkgReader = &buf

		err := compareChecksum(d.Logger, r, verifiedChecksum)
		if err != nil {
			return err
		}
	}

	pkgFile, err := ioutil.TempFile("", pb.Filename)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	d.Logger.Printf("copying downloaded file to %s", pkgFile.Name())
	bytesCopied, err := io.Copy(pkgFile, pkgReader)
	if err != nil {
		return err
	}
	d.Logger.Printf("copied %d bytes to %s", bytesCopied, pkgFile.Name())

	expectedSize := 0
	if length := resp.Header.Get("content-length"); length != "" {
		var err error
		expectedSize, err = strconv.Atoi(length)
		if err != nil {
			return err
		}
	}
	if expectedSize != 0 && bytesCopied != int64(expectedSize) {
		return fmt.Errorf("unexpected size (downloaded: %d, expected: %d)",
			bytesCopied, expectedSize)
	}

	r, err := zip.OpenReader(pkgFile.Name())
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		srcFile, err := f.Open()
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, f.Name)
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		srcFile.Close()
		dstFile.Close()
	}

	return nil
}
