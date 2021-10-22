package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	isrc "github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/validators"
	"github.com/hashicorp/hc-install/product"
)

var (
	cloneTimeout  = 1 * time.Minute
	buildTimeout  = 2 * time.Minute
	discardLogger = log.New(ioutil.Discard, "", 0)
)

// GitRevision installs a particular git revision by cloning
// the repository and building it per product BuildInstructions
type GitRevision struct {
	Product      product.Product
	InstallDir   string
	Ref          string
	CloneTimeout time.Duration
	BuildTimeout time.Duration

	logger        *log.Logger
	pathsToRemove []string
}

func (*GitRevision) IsSourceImpl() isrc.InstallSrcSigil {
	return isrc.InstallSrcSigil{}
}

func (gr *GitRevision) SetLogger(logger *log.Logger) {
	gr.logger = logger
}

func (gr *GitRevision) log() *log.Logger {
	if gr.logger == nil {
		return discardLogger
	}
	return gr.logger
}

func (gr *GitRevision) Validate() error {
	if !validators.IsProductNameValid(gr.Product.Name) {
		return fmt.Errorf("invalid product name: %q", gr.Product.Name)
	}
	if !validators.IsBinaryNameValid(gr.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", gr.Product.BinaryName())
	}

	bi := gr.Product.BuildInstructions
	if bi == nil {
		return fmt.Errorf("no build instructions")
	}
	if bi.GitRepoURL == "" {
		return fmt.Errorf("missing repository URL")
	}
	if bi.Build == nil {
		return fmt.Errorf("missing build instructions")
	}

	return nil
}

func (gr *GitRevision) Build(ctx context.Context) (string, error) {
	buildTimeout := buildTimeout
	if gr.BuildTimeout > 0 {
		buildTimeout = gr.BuildTimeout
	}

	bi := gr.Product.BuildInstructions

	if bi.PreCloneCheck != nil {
		pccCtx, cancelFunc := context.WithTimeout(ctx, buildTimeout)
		defer cancelFunc()

		gr.log().Printf("running pre-clone check (timeout: %s)", buildTimeout)
		err := bi.PreCloneCheck.Check(pccCtx)
		if err != nil {
			return "", err
		}
		gr.log().Printf("pre-clone check finished")
	}

	if gr.pathsToRemove == nil {
		gr.pathsToRemove = make([]string, 0)
	}

	repoDir, err := ioutil.TempDir("",
		fmt.Sprintf("hc-install-build-%s", gr.Product.Name))
	if err != nil {
		return "", err
	}
	gr.pathsToRemove = append(gr.pathsToRemove, repoDir)

	ref := gr.Ref
	if ref == "" {
		ref = "HEAD"
	}

	timeout := cloneTimeout
	if gr.BuildTimeout > 0 {
		timeout = gr.BuildTimeout
	}
	cloneCtx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	gr.log().Printf("cloning repository from %s to %s (timeout: %s)",
		gr.Product.BuildInstructions.GitRepoURL,
		repoDir, timeout)
	repo, err := git.PlainCloneContext(cloneCtx, repoDir, false, &git.CloneOptions{
		URL:           gr.Product.BuildInstructions.GitRepoURL,
		ReferenceName: plumbing.ReferenceName(gr.Ref),
		Depth:         1,
	})
	if err != nil {
		return "", fmt.Errorf("unable to clone %q: %w",
			gr.Product.BuildInstructions.GitRepoURL, err)
	}
	gr.log().Printf("cloning finished")
	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	gr.log().Printf("repository HEAD is at %s", head.Hash())

	buildCtx, cancelFunc := context.WithTimeout(ctx, buildTimeout)
	defer cancelFunc()

	if loggableBuilder, ok := bi.Build.(withLogger); ok {
		loggableBuilder.SetLogger(gr.log())
	}
	installDir := gr.InstallDir
	if installDir == "" {
		tmpDir, err := ioutil.TempDir("",
			fmt.Sprintf("hc-install-%s-%s", gr.Product.Name, head.Hash()))
		if err != nil {
			return "", err
		}
		installDir = tmpDir
		gr.pathsToRemove = append(gr.pathsToRemove, installDir)
	}

	gr.log().Printf("building (timeout: %s)", buildTimeout)
	return bi.Build.Build(buildCtx, repoDir, installDir, gr.Product.BinaryName())
}

func (gr *GitRevision) Remove(ctx context.Context) error {
	if gr.pathsToRemove != nil {
		for _, path := range gr.pathsToRemove {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
	}

	return gr.Product.BuildInstructions.Build.Remove(ctx)
}

type withLogger interface {
	SetLogger(*log.Logger)
}
