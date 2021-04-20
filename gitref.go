package hcinstall

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GitRef(ref string) *GitGetter {
	return &GitGetter{
		ref: ref,
	}
}

func GitCommit(hash string) *GitGetter {
	return &GitGetter{
		commit: hash,
	}
}

type GitGetter struct {
	getter
	ref    string
	commit string
}

func (g *GitGetter) Get(ctx context.Context) (string, error) {
	if g.c.Product.RepoURL == "" {
		return "", fmt.Errorf("GitRefGetter is not available for product %s", g.c.Product.Name)
	}

	tmpBuildDir, err := ioutil.TempDir("", "hcinstall-build")
	if err != nil {
		return "", err
	}

	if g.ref != "" {
		ref := plumbing.ReferenceName(g.ref)
		_, err := git.PlainClone(tmpBuildDir, false, &git.CloneOptions{
			URL:           g.c.Product.RepoURL,
			ReferenceName: ref,
			Depth:         1,
			Tags:          git.NoTags,
		})
		if err != nil {
			return "", fmt.Errorf("Unable to clone %q: %w", g.c.Product.RepoURL, err)
		}
	} else if g.commit != "" {
		repo, err := git.PlainClone(tmpBuildDir, false, &git.CloneOptions{
			URL:  g.c.Product.RepoURL,
			Tags: git.NoTags,
		})
		worktree, err := repo.Worktree()
		if err != nil {
			return "", fmt.Errorf("Error obtaining worktree: %w", err)
		}

		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(g.commit),
		})
		if err != nil {
			return "", fmt.Errorf("Error checking out commit %s: %w", g.commit, err)
		}

	} else {
		return "", errors.New("Either ref or commit must be specified for GitGetter. Please use GitRef() or GitCommit() functions.")
	}

	var productFilename string
	if runtime.GOOS == "windows" {
		productFilename = g.c.Product.Name + ".exe"
	} else {
		productFilename = g.c.Product.Name
	}

	goArgs := []string{"build", "-o", filepath.Join(g.c.InstallDir, productFilename)}

	// TODO is this needed?
	vendorDir := filepath.Join(g.c.InstallDir, "vendor")
	if fi, err := os.Stat(vendorDir); err == nil && fi.IsDir() {
		goArgs = append(goArgs, "-mod", "vendor")
	}

	cmd := exec.CommandContext(ctx, "go", goArgs...)
	cmd.Dir = tmpBuildDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to build Terraform: %w\n%s", err, out)
	}

	return filepath.Join(g.c.InstallDir, productFilename), nil
}
