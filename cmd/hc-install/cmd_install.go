package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
	"gophers.dev/pkgs/ignore"

	hci "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

type InstallCommand struct {
	Ui cli.Ui
}

func (c *InstallCommand) Name() string { return "install" }

func (c *InstallCommand) Synopsis() string {
	return "Install a HashiCorp product"
}

func (c *InstallCommand) Help() string {
	helpText := `
Usage: hc-install install [options] -version <version> <product>

  This command installs a HashiCorp product.
  Options:
    -version  [REQUIRED] Version of product to install.
    -path     Path to directory where the product will be installed. Defaults
              to current working directory.
`
	return strings.TrimSpace(helpText)
}

func (c *InstallCommand) Run(args []string) int {

	var (
		version        string
		installDirPath string
	)

	fs := flag.NewFlagSet("install", flag.ExitOnError)
	fs.Usage = func() { c.Ui.Output(c.Help()) }
	fs.StringVar(&version, "version", "", "version of product to install")
	fs.StringVar(&installDirPath, "path", "", "path to directory where production will be installed")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// golang's arg parser is Posix-compliant but doesn't match the
	// common GNU flag parsing argument, so force an error rather than
	// silently dropping the options
	args = fs.Args()
	if len(args) != 1 {
		c.Ui.Error(`This command requires one positional argument: <product>
Option flags must be provided before the positional argument`)
		return 1
	}
	product := fs.Args()[0]

	if version == "" {
		c.Ui.Error("-version flag is required")
		return 1
	}

	if installDirPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Could not get current working directory for default installation path: %v", err))
		}
		installDirPath = cwd
		return 1
	}

	if err := c.install(product, version, installDirPath); err != nil {
		msg := fmt.Sprintf("failed to install %s@%s: %v", product, version, err)
		c.Ui.Error(msg)
		return 1
	}

	return 0
}

func (c *InstallCommand) install(project, tag, installDirPath string) error {
	msg := fmt.Sprintf("hc-install: will install %s@%s", project, tag)
	c.Ui.Info(msg)

	v, err := version.NewVersion(tag)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	i := hci.NewInstaller()

	source := &releases.ExactVersion{
		Product: product.Product{
			Name: project,
			BinaryName: func() string {
				if runtime.GOOS == "windows" {
					return fmt.Sprintf("%s.exe", project)
				}
				return project
			},
		},
		Version:    v,
		InstallDir: installDirPath,
	}

	ctx := context.Background()
	_, err = i.Install(ctx, []src.Installable{source})
	return err
}

func (c *InstallCommand) copyProgram(programPath, installDirPath string) error {
	program := filepath.Base(programPath)
	destination := filepath.Join(installDirPath, program)

	msg := fmt.Sprintf("hc-install: copy executable to %s", destination)
	c.Ui.Info(msg)

	return clone(programPath, destination)
}

func clone(source, destination string) error {
	sFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer ignore.Close(sFile)

	dFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer ignore.Close(dFile)

	_, err = io.Copy(dFile, sFile)
	return err
}
