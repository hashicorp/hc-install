// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package install

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hc-install/errors"
	"github.com/hashicorp/hc-install/src"
)

type Installer struct {
	logger *log.Logger

	removableSources []src.Removable
}

type RemoveFunc func(ctx context.Context) error

func NewInstaller() *Installer {
	discardLogger := log.New(io.Discard, "", 0)
	return &Installer{
		logger: discardLogger,
	}
}

func (i *Installer) SetLogger(logger *log.Logger) {
	i.logger = logger
}

func (i *Installer) Ensure(ctx context.Context, sources []src.Source) (*src.Details, error) {
	var errs *multierror.Error

	for _, source := range sources {
		if srcWithLogger, ok := source.(src.LoggerSettable); ok {
			srcWithLogger.SetLogger(i.logger)
		}

		if srcValidatable, ok := source.(src.Validatable); ok {
			err := srcValidatable.Validate()
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	if errs.ErrorOrNil() != nil {
		return nil, errs
	}

	i.removableSources = make([]src.Removable, 0)

	for _, source := range sources {
		if s, ok := source.(src.Removable); ok {
			i.removableSources = append(i.removableSources, s)
		}

		switch s := source.(type) {
		case src.Findable:
			details, err := s.Find(ctx)
			if err != nil {
				if errors.IsErrorSkippable(err) {
					errs = multierror.Append(errs, err)
					continue
				}
				return nil, err
			}

			return details, nil
		case src.Installable:
			details, err := s.Install(ctx)
			if err != nil {
				if errors.IsErrorSkippable(err) {
					errs = multierror.Append(errs, err)
					continue
				}
				return nil, err
			}

			return details, nil
		case src.Buildable:
			execPath, err := s.Build(ctx)
			if err != nil {
				if errors.IsErrorSkippable(err) {
					errs = multierror.Append(errs, err)
					continue
				}
				return nil, err
			}

			return &src.Details{
				ExecutablePath: execPath,
				Version:        nil,
			}, nil
		default:
			return nil, fmt.Errorf("unknown source: %T", s)
		}
	}

	return nil, fmt.Errorf("unable to find, install, or build from %d sources: %s",
		len(sources), errs.ErrorOrNil())
}

func (i *Installer) Install(ctx context.Context, sources []src.Installable) (*src.Details, error) {
	var errs *multierror.Error

	i.removableSources = make([]src.Removable, 0)

	for _, source := range sources {
		if srcWithLogger, ok := source.(src.LoggerSettable); ok {
			srcWithLogger.SetLogger(i.logger)
		}

		if srcValidatable, ok := source.(src.Validatable); ok {
			err := srcValidatable.Validate()
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
		}

		if s, ok := source.(src.Removable); ok {
			i.removableSources = append(i.removableSources, s)
		}

		installDetails, err := source.Install(ctx)
		if err != nil {
			if errors.IsErrorSkippable(err) {
				errs = multierror.Append(errs, err)
				continue
			}
			return nil, err
		}

		result := &src.Details{
			Version:        installDetails.Version,
			ExecutablePath: installDetails.ExecutablePath,
		}
		return result, nil
	}

	return nil, fmt.Errorf("unable install from %d sources: %s",
		len(sources), errs.ErrorOrNil())
}

func (i *Installer) Remove(ctx context.Context) error {
	var errs *multierror.Error

	if i.removableSources != nil {
		for _, rs := range i.removableSources {
			err := rs.Remove(ctx)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs.ErrorOrNil()
}
