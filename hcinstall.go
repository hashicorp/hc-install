package hcinstall

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcinstall/products"
)

type Getter interface {
	Get(context.Context) (string, error)
	SetClient(*Client)
}

type getter struct {
	c                  *Client
	VersionConstraints VersionConstraints
}

func (g *getter) SetClient(c *Client) { g.c = c }

// Client is a client for finding and installing binaries.
//
// Convenience functions such as hcinstall.Install use a Client with default
// values. A Client can be instantiated and
type Client struct {
	Product products.Product

	InstallDir string

	Getters []Getter

	VersionConstraints *VersionConstraints

	ForceCheckpoint bool
}

func (c *Client) Install(ctx context.Context) (string, error) {
	var execPath string

	for _, getter := range c.Getters {
		getter.SetClient(c)
	}

	// go through the options in order
	// until a valid terraform executable is found
	for _, g := range c.Getters {
		p, err := g.Get(ctx)
		if err != nil {
			return "", fmt.Errorf("unexpected error: %s", err)
		}

		// assert version
		if err := c.assertVersion(p); err != nil {
			log.Printf("[WARN] Executable at %s did not satisfy version constraint: %s", p, err)
			continue
		}

		if p == "" {
			// strategy did not locate an executable - fall through to next
			continue
		} else {
			execPath = p
			break
		}
	}

	if execPath == "" {
		return "", fmt.Errorf("could not find executable")
	}

	return execPath, nil
}

func (c *Client) assertVersion(execPath string) error {
	var v *version.Version
	if c.VersionConstraints.latest {
		resp, err := checkpoint.Check(&checkpoint.CheckParams{
			Product: c.Product.Name,
			Force:   c.ForceCheckpoint,
		})
		if err != nil {
			return err
		}

		if resp.CurrentVersion == "" {
			return fmt.Errorf("could not determine latest version of terraform using checkpoint: CHECKPOINT_DISABLE may be set")
		}

		v, err = version.NewVersion(resp.CurrentVersion)
		if err != nil {
			return err
		}
	} else {
		v = c.VersionConstraints.exact
	}

	actualVersion, err := c.Product.GetVersion(execPath)
	if err != nil {
		return err
	}

	if !actualVersion.Equal(v) {
		return fmt.Errorf("reported version %s did not match required version %s", actualVersion, v)
	}

	return nil
}

// Install downloads and verifies the signature of the specified product
// executable, returning its path.
// Note that the DefaultFinders are applied in order, and therefore if a local
// executable is found that satisfies the version constraints and checksum,
// no download need take place.
func Install(ctx context.Context, dstDir string, product products.Product, versionConstraints string) (string, error) {
	installDir, err := ensureInstallDir(dstDir)
	if err != nil {
		return "", err
	}

	v, err := NewVersionConstraints(versionConstraints)
	if err != nil {
		return "", err
	}

	defaultGetters := []Getter{LookPath(), Releases()}

	c := Client{
		InstallDir:         installDir,
		Getters:            defaultGetters,
		VersionConstraints: v,
		Product:            product,
	}

	return c.Install(ctx)
}

// ensureInstallDir checks whether the supplied installDir is suitable for the
// downloaded binary, creating a temporary directory if installDir is blank.
func ensureInstallDir(installDir string) (string, error) {
	if installDir == "" {
		return ioutil.TempDir("", "hcinstall")
	}

	if _, err := os.Stat(installDir); err != nil {
		return "", fmt.Errorf("could not access directory %s for installation: %w", installDir, err)
	}

	return installDir, nil
}
