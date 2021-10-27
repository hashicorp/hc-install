package fs

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/hc-install/errors"
	"github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/validators"
	"github.com/hashicorp/hc-install/product"
)

// AnyVersion finds the first executable binary of the product name
// within system $PATH and any declared ExtraPaths
// (which are *appended* to any directories in $PATH)
type AnyVersion struct {
	Product    product.Product
	ExtraPaths []string

	logger *log.Logger
}

func (*AnyVersion) IsSourceImpl() src.InstallSrcSigil {
	return src.InstallSrcSigil{}
}

func (av *AnyVersion) Validate() error {
	if !validators.IsBinaryNameValid(av.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", av.Product.BinaryName())
	}
	return nil
}

func (av *AnyVersion) SetLogger(logger *log.Logger) {
	av.logger = logger
}

func (av *AnyVersion) log() *log.Logger {
	if av.logger == nil {
		return discardLogger
	}
	return av.logger
}

func (av *AnyVersion) Find(ctx context.Context) (string, error) {
	execPath, err := findFile(lookupDirs(av.ExtraPaths), av.Product.BinaryName(), checkExecutable)
	if err != nil {
		return "", errors.SkippableErr(err)
	}

	if !filepath.IsAbs(execPath) {
		var err error
		execPath, err = filepath.Abs(execPath)
		if err != nil {
			return "", errors.SkippableErr(err)
		}
	}
	return execPath, nil
}
