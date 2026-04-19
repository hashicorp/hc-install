// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package releasesjson

import (
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/hashicorp/hc-install/internal/pubkey"
)

// Verifies that when DefaultPublicKey contains multiple concatenated armored
// blocks sharing a primary key, keyEntityList reads all of them and keeps the
// newest identity self-signature — letting a refreshed block extend validity
// of the original key.
func TestChecksumDownloader_keyEntityList_refreshedSelfSignature(t *testing.T) {
	blocks := strings.Count(pubkey.DefaultPublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----")
	if blocks < 2 {
		t.Skipf("DefaultPublicKey contains %d armored block(s); test requires >= 2", blocks)
	}

	baseline, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pubkey.DefaultPublicKey))
	if err != nil {
		t.Fatalf("baseline ReadArmoredKeyRing: %v", err)
	}
	if len(baseline) != 1 {
		t.Fatalf("baseline expected 1 entity, got %d", len(baseline))
	}
	baselineSelfSig := latestSelfSig(baseline[0])

	cd := &ChecksumDownloader{ArmoredPublicKey: pubkey.DefaultPublicKey}
	entities, err := cd.keyEntityList()
	if err != nil {
		t.Fatalf("keyEntityList: %v", err)
	}
	if len(entities) == 0 {
		t.Fatal("no entities returned")
	}
	fixedSelfSig := latestSelfSig(entities[0])

	if !fixedSelfSig.After(baselineSelfSig) {
		t.Fatalf("expected a newer self-signature than baseline (%s); got %s",
			baselineSelfSig, fixedSelfSig)
	}
}

func latestSelfSig(e *openpgp.Entity) time.Time {
	var latest time.Time
	for _, id := range e.Identities {
		if id.SelfSignature != nil && id.SelfSignature.CreationTime.After(latest) {
			latest = id.SelfSignature.CreationTime
		}
	}
	return latest
}

func TestChecksumDownloader_keyEntityList_empty(t *testing.T) {
	cd := &ChecksumDownloader{}
	if _, err := cd.keyEntityList(); err == nil {
		t.Fatal("expected error for empty ArmoredPublicKey, got nil")
	}
}

func TestChecksumDownloader_keyEntityList_noBlocks(t *testing.T) {
	cd := &ChecksumDownloader{ArmoredPublicKey: "not an armored block"}
	_, err := cd.keyEntityList()
	if err == nil {
		t.Fatal("expected error for input with no armored blocks, got nil")
	}
	if !strings.Contains(err.Error(), "no keys found") {
		t.Fatalf("expected 'no keys found' error, got: %v", err)
	}
}

func TestChecksumDownloader_keyEntityList_malformedBlock(t *testing.T) {
	const malformed = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n!!!not base64!!!\n-----END PGP PUBLIC KEY BLOCK-----"
	cd := &ChecksumDownloader{ArmoredPublicKey: malformed}
	if _, err := cd.keyEntityList(); err == nil {
		t.Fatal("expected error for malformed armored block, got nil")
	}
}

func TestSplitArmoredBlocks(t *testing.T) {
	const blockA = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\naaaa\n-----END PGP PUBLIC KEY BLOCK-----"
	const blockB = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nbbbb\n-----END PGP PUBLIC KEY BLOCK-----"

	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "empty",
			in:   "",
			want: nil,
		},
		{
			name: "single block",
			in:   blockA,
			want: []string{blockA},
		},
		{
			name: "two blocks back to back",
			in:   blockA + "\n" + blockB,
			want: []string{blockA, blockB},
		},
		{
			name: "leading and trailing garbage",
			in:   "junk before\n" + blockA + "\njunk after\n",
			want: []string{blockA},
		},
		{
			name: "missing end marker",
			in:   "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\naaaa\n",
			want: nil,
		},
		{
			name: "no begin marker",
			in:   "just some text\nwith no armor\n",
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitArmoredBlocks(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d blocks, want %d: %q", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("block %d:\n got: %q\nwant: %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
