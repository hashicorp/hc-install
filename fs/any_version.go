package fs

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/magodo/hc-install/errors"
	"github.com/magodo/hc-install/internal/src"
	"github.com/magodo/hc-install/internal/validators"
	"github.com/magodo/hc-install/product"
)

// AnyVersion finds an executable binary of any version
// either defined by ExactBinPath, or as part of Product.
//
// When ExactBinPath is used, the source is skipped when
// the binary is not found or accessible/executable.
//
// When Product is used, binary name is looked up within system $PATH
// and any declared ExtraPaths (which are *appended* to
// any directories in $PATH). Source is skipped if no binary
// is found or accessible/executable.
//
// When Constraints is used, find the first binary that meets the specified
// version constraint.
type AnyVersion struct {
	// Product represents the product (its binary name to look up),
	// conflicts with ExactBinPath
	Product *product.Product

	// ExtraPaths represents additional dir paths to be appended to
	// the default system $PATH, conflicts with ExactBinPath
	ExtraPaths []string

	// Constraints represents a set of version constraints that should
	// be met by the binary, conflicts with ExactBinPath
	Constraints version.Constraints

	// ExactBinPath represents exact path to the binary,
	// conflicts with Product, ExtraPaths and Constraints
	ExactBinPath string

	logger *log.Logger
}

func (*AnyVersion) IsSourceImpl() src.InstallSrcSigil {
	return src.InstallSrcSigil{}
}

func (av *AnyVersion) Validate() error {
	if av.ExactBinPath == "" && av.Product == nil {
		return fmt.Errorf("must use either ExactBinPath or Product + ExtraPaths")
	}
	if av.ExactBinPath != "" && (av.Product != nil || len(av.ExtraPaths) > 0) {
		return fmt.Errorf("use either ExactBinPath or Product + ExtraPaths, not both")
	}
	if av.ExactBinPath != "" && !filepath.IsAbs(av.ExactBinPath) {
		return fmt.Errorf("expected ExactBinPath (%q) to be an absolute path", av.ExactBinPath)
	}
	if av.Product != nil && !validators.IsBinaryNameValid(av.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", av.Product.BinaryName())
	}
	if av.Constraints != nil {
		if av.Product.GetVersion == nil {
			return fmt.Errorf("undeclared version getter")
		}
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
	if av.ExactBinPath != "" {
		err := checkExecutable(av.ExactBinPath)
		if err != nil {
			return "", errors.SkippableErr(err)
		}

		return av.ExactBinPath, nil
	}

	execPath, err := findFile(lookupDirs(av.ExtraPaths), av.Product.BinaryName(), func(file string) error {
		err := checkExecutable(file)
		if err != nil {
			return err
		}

		if av.Constraints == nil {
			return nil
		}

		v, err := av.Product.GetVersion(ctx, file)
		if err != nil {
			return err
		}

		for _, vc := range av.Constraints {
			if !vc.Check(v) {
				return fmt.Errorf("version (%s) doesn't meet constraint %s", v, vc.String())
			}
		}

		return nil
	})
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
