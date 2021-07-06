package releasesjson

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hashicorp/hc-install/internal/httpclient"
	"golang.org/x/crypto/openpgp"
)

type ChecksumDownloader struct {
	ProductVersion *ProductVersion
	Logger         *log.Logger
}

type ChecksumFileMap map[string]HashSum

type HashSum []byte

func (hs HashSum) Size() int {
	return len(hs)
}

func (hs HashSum) String() string {
	return hex.EncodeToString(hs)
}

func HashSumFromHexDigest(hexDigest string) (HashSum, error) {
	sumBytes, err := hex.DecodeString(hexDigest)
	if err != nil {
		return nil, err
	}
	return HashSum(sumBytes), nil
}

func (cd *ChecksumDownloader) DownloadAndVerifyChecksums() (ChecksumFileMap, error) {
	sigFilename, err := findSigFilename(cd.ProductVersion)
	if err != nil {
		return nil, err
	}

	client := httpclient.NewHTTPClient()
	sigURL := fmt.Sprintf("%s/%s/%s/%s", baseURL,
		cd.ProductVersion.Name,
		cd.ProductVersion.Version,
		sigFilename)
	sigResp, err := client.Get(sigURL)
	if err != nil {
		return nil, err
	}
	defer sigResp.Body.Close()

	shasumsURL := fmt.Sprintf("%s/%s/%s/%s", baseURL,
		cd.ProductVersion.Name,
		cd.ProductVersion.Version,
		cd.ProductVersion.SHASUMS)
	sumsResp, err := client.Get(shasumsURL)
	if err != nil {
		return nil, err
	}
	defer sumsResp.Body.Close()

	var shaSums strings.Builder
	sumsReader := io.TeeReader(sumsResp.Body, &shaSums)

	err = verifySumsSignature(sumsReader, sigResp.Body)
	if err != nil {
		return nil, err
	}

	return fileMapFromChecksums(shaSums)
}

func fileMapFromChecksums(checksums strings.Builder) (ChecksumFileMap, error) {
	csMap := make(ChecksumFileMap, 0)

	lines := strings.Split(checksums.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected checksum line format: %q", line)
		}

		h, err := HashSumFromHexDigest(parts[0])
		if err != nil {
			return nil, err
		}

		if h.Size() != sha256.Size {
			return nil, fmt.Errorf("unexpected sha256 format (len: %d, expected: %d)",
				h.Size(), sha256.Size)
		}

		csMap[parts[1]] = h
	}
	return csMap, nil
}

func compareChecksum(logger *log.Logger, r io.Reader, verifiedHashSum HashSum) error {
	h := sha256.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return err
	}

	calculatedSum := h.Sum(nil)
	if !bytes.Equal(calculatedSum, verifiedHashSum) {
		return fmt.Errorf("checksum mismatch (expected %q, calculated %q)",
			verifiedHashSum,
			hex.EncodeToString(calculatedSum))
	}

	logger.Printf("checksum matches: %q", hex.EncodeToString(calculatedSum))

	return nil
}

func verifySumsSignature(checksums, signature io.Reader) error {
	el, err := openpgp.ReadArmoredKeyRing(strings.NewReader(publicKey))
	if err != nil {
		return err
	}

	_, err = openpgp.CheckDetachedSignature(el, checksums, signature)

	return err
}

func findSigFilename(pv *ProductVersion) (string, error) {
	sigFiles := pv.SHASUMSSigs
	if len(sigFiles) == 0 {
		sigFiles = []string{pv.SHASUMSSig}
	}

	for _, filename := range sigFiles {
		if strings.HasSuffix(filename, fmt.Sprintf("_SHA256SUMS.%s.sig", keyID)) {
			return filename, nil
		}
		if strings.HasSuffix(filename, "_SHA256SUMS.sig") {
			return filename, nil
		}
	}

	return "", fmt.Errorf("no suitable sig file found")
}
