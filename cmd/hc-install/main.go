package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	hci "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/logutils"
	"github.com/mitchellh/cli"
	"gophers.dev/pkgs/extractors/env"
	"gophers.dev/pkgs/ignore"
)

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	ui := &cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}

	exitStatus := run(ui, os.Args[1:])

	os.Exit(exitStatus)
}

func run(ui cli.Ui, args []string) int {
	if len(args) != 2 {
		ui.Error("usage: hc-install PRODUCT VERSION")
		return 1
	}

	if err := install(ui, args[0], args[1]); err != nil {
		msg := fmt.Sprintf("failed to install %s@%s: %v", args[0], args[1], err)
		ui.Error(msg)
		return 1
	}

	return 0
}

func install(ui cli.Ui, project, tag string) error {
	msg := fmt.Sprintf("hc-install: will install %s@%s", project, tag)
	ui.Info(msg)

	v := version.Must(version.NewVersion(tag))
	i := hci.NewInstaller()

	var source src.Source
	switch project {
	case "consul":
		source = &releases.ExactVersion{
			Product: product.Consul,
			Version: v,
		}
	case "vault":
		source = &releases.ExactVersion{
			Product: product.Vault,
			Version: v,
		}
	default:
		return fmt.Errorf("project %s cannot be downloaded", project)
	}

	ctx := context.Background()
	executable, instErr := i.Ensure(ctx, []src.Source{source})
	if instErr != nil {
		return instErr
	}

	if err := copyProgram(ui, executable); err != nil {
		return err
	}

	return nil
}

func copyProgram(ui cli.Ui, programPath string) error {
	var (
		gobin   string
		gopath  string
		program = filepath.Base(programPath)
	)
	if err := env.ParseOS(env.Schema{
		"GOBIN":  env.String(&gobin, false),
		"GOPATH": env.String(&gopath, false),
	}); err != nil {
		return err
	}

	var destination string
	switch {
	case gobin != "":
		destination = filepath.Join(gobin, program)
	case gopath != "":
		destination = filepath.Join(gopath, "bin", program)
	}

	msg := fmt.Sprintf("hc-install: copy executable to %s", destination)
	ui.Info(msg)

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
