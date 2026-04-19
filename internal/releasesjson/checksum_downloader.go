// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package releasesjson

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/hashicorp/hc-install/internal/httpclient"
)

type ChecksumDownloader struct {
	ProductVersion   *ProductVersion
	Logger           *log.Logger
	ArmoredPublicKey string

	BaseURL string
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

func (cd *ChecksumDownloader) DownloadAndVerifyChecksums(ctx context.Context) (ChecksumFileMap, error) {
	sigFilename, err := cd.findSigFilename(cd.ProductVersion)
	if err != nil {
		return nil, err
	}

	client := httpclient.NewHTTPClient(cd.Logger)
	sigURL := fmt.Sprintf("%s/%s/%s/%s", cd.BaseURL,
		url.PathEscape(cd.ProductVersion.Name),
		url.PathEscape(cd.ProductVersion.Version.String()),
		url.PathEscape(sigFilename))
	cd.Logger.Printf("downloading signature from %s", sigURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sigURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %q: %w", sigURL, err)
	}
	sigResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if sigResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download signature from %q: %s", sigURL, sigResp.Status)
	}

	defer sigResp.Body.Close()

	shasumsURL := fmt.Sprintf("%s/%s/%s/%s", cd.BaseURL,
		url.PathEscape(cd.ProductVersion.Name),
		url.PathEscape(cd.ProductVersion.Version.String()),
		url.PathEscape(cd.ProductVersion.SHASUMS))
	cd.Logger.Printf("downloading checksums from %s", shasumsURL)

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, shasumsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %q: %w", shasumsURL, err)
	}
	sumsResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if sumsResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download checksums from %q: %s", shasumsURL, sumsResp.Status)
	}

	defer sumsResp.Body.Close()

	var shaSums strings.Builder
	sumsReader := io.TeeReader(sumsResp.Body, &shaSums)

	err = cd.verifySumsSignature(sumsReader, sigResp.Body)
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

func (cd *ChecksumDownloader) verifySumsSignature(checksums, signature io.Reader) error {
	el, err := cd.keyEntityList()
	if err != nil {
		return err
	}

	_, err = openpgp.CheckDetachedSignature(el, checksums, signature, nil)
	if err != nil {
		return fmt.Errorf("unable to verify checksums signature: %w", err)
	}

	cd.Logger.Printf("checksum signature is valid")

	return nil
}

func (cd *ChecksumDownloader) findSigFilename(pv *ProductVersion) (string, error) {
	sigFiles := pv.SHASUMSSigs
	if len(sigFiles) == 0 {
		sigFiles = []string{pv.SHASUMSSig}
	}

	keyIds, err := cd.pubKeyIds()
	if err != nil {
		return "", err
	}

	for _, filename := range sigFiles {
		for _, keyID := range keyIds {
			if strings.HasSuffix(filename, fmt.Sprintf("_SHA256SUMS.%s.sig", keyID)) {
				return filename, nil
			}
		}
		if strings.HasSuffix(filename, "_SHA256SUMS.sig") {
			return filename, nil
		}
	}

	return "", fmt.Errorf("no suitable sig file found")
}

func (cd *ChecksumDownloader) pubKeyIds() ([]string, error) {
	entityList, err := cd.keyEntityList()
	if err != nil {
		return nil, err
	}

	fingerprints := make([]string, 0)
	for _, entity := range entityList {
		fingerprints = append(fingerprints, entity.PrimaryKey.KeyIdShortString())
	}

	return fingerprints, nil
}

func (cd *ChecksumDownloader) keyEntityList() (openpgp.EntityList, error) {
	if cd.ArmoredPublicKey == "" {
		return nil, fmt.Errorf("no public key provided")
	}
	// ArmoredPublicKey may contain more than one concatenated armored block
	// (e.g. an original key plus a block with refreshed self-signatures).
	// openpgp.ReadArmoredKeyRing only decodes the first block, so split the
	// input, parse each block independently, and merge entities that share a
	// primary key so the newest self-signatures win.
	var entities openpgp.EntityList
	for _, block := range splitArmoredBlocks(cd.ArmoredPublicKey) {
		part, err := openpgp.ReadArmoredKeyRing(strings.NewReader(block))
		if err != nil {
			return nil, err
		}
		entities = append(entities, part...)
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("no keys found in armored data")
	}
	return mergeEntitiesByPrimaryKey(entities), nil
}

// mergeEntitiesByPrimaryKey collapses entities that share a primary key into
// one, keeping the most recent identity self-signature and subkey binding
// signature. This lets a refreshed armored block supersede the self-signatures
// of an older block that repeats the same primary key.
func mergeEntitiesByPrimaryKey(entities openpgp.EntityList) openpgp.EntityList {
	merged := make(map[string]*openpgp.Entity, len(entities))
	order := make([]string, 0, len(entities))
	for _, e := range entities {
		fpr := string(e.PrimaryKey.Fingerprint)
		existing, ok := merged[fpr]
		if !ok {
			merged[fpr] = e
			order = append(order, fpr)
			continue
		}
		mergeEntity(existing, e)
	}
	out := make(openpgp.EntityList, 0, len(order))
	for _, fpr := range order {
		out = append(out, merged[fpr])
	}
	return out
}

func mergeEntity(dst, src *openpgp.Entity) {
	for name, srcID := range src.Identities {
		dstID, ok := dst.Identities[name]
		if !ok {
			dst.Identities[name] = srcID
			continue
		}
		if srcID.SelfSignature != nil &&
			(dstID.SelfSignature == nil ||
				srcID.SelfSignature.CreationTime.After(dstID.SelfSignature.CreationTime)) {
			dstID.SelfSignature = srcID.SelfSignature
		}
		dstID.Signatures = append(dstID.Signatures, srcID.Signatures...)
		dstID.Revocations = append(dstID.Revocations, srcID.Revocations...)
	}
	for _, srcSK := range src.Subkeys {
		matched := false
		for i := range dst.Subkeys {
			if dst.Subkeys[i].PublicKey.KeyId != srcSK.PublicKey.KeyId {
				continue
			}
			matched = true
			if srcSK.Sig != nil &&
				(dst.Subkeys[i].Sig == nil ||
					srcSK.Sig.CreationTime.After(dst.Subkeys[i].Sig.CreationTime)) {
				dst.Subkeys[i].Sig = srcSK.Sig
			}
			dst.Subkeys[i].Revocations = append(dst.Subkeys[i].Revocations, srcSK.Revocations...)
			break
		}
		if !matched {
			dst.Subkeys = append(dst.Subkeys, srcSK)
		}
	}
	dst.Revocations = append(dst.Revocations, src.Revocations...)
}

// splitArmoredBlocks returns each ASCII-armored block in s as a separate
// string. It tolerates leading/trailing whitespace and ignores any content
// outside BEGIN/END markers.
func splitArmoredBlocks(s string) []string {
	const beginMarker = "-----BEGIN "
	const endMarker = "-----END "
	const endOfLine = "-----"

	var blocks []string
	for {
		beginIdx := strings.Index(s, beginMarker)
		if beginIdx == -1 {
			break
		}
		rest := s[beginIdx:]
		// Find the end marker after the begin.
		endIdx := strings.Index(rest, endMarker)
		if endIdx == -1 {
			break
		}
		// Advance past the end marker line (to the next "-----" closing the line).
		tail := rest[endIdx+len(endMarker):]
		closeIdx := strings.Index(tail, endOfLine)
		if closeIdx == -1 {
			break
		}
		blockEnd := endIdx + len(endMarker) + closeIdx + len(endOfLine)
		blocks = append(blocks, rest[:blockEnd])
		s = rest[blockEnd:]
	}
	return blocks
}
