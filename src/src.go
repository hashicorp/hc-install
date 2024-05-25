// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package src

import (
	"context"
	"log"

	"github.com/hashicorp/go-version"
	isrc "github.com/hashicorp/hc-install/internal/src"
)

// Source represents an installer, finder, or builder
type Source interface {
	IsSourceImpl() isrc.InstallSrcSigil
}

type Details struct {
	Version        *version.Version
	ExecutablePath string
}

type Installable interface {
	Source
	Install(ctx context.Context) (*Details, error)
}

type Findable interface {
	Source
	Find(ctx context.Context) (*Details, error)
}

type Buildable interface {
	Source
	Build(ctx context.Context) (string, error)
}

type Validatable interface {
	Source
	Validate() error
}

type Removable interface {
	Source
	Remove(ctx context.Context) error
}

type LoggerSettable interface {
	SetLogger(logger *log.Logger)
}
