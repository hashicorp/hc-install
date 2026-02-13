// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"log"
	"os"

	"github.com/hashicorp/hc-install/version"

	"github.com/hashicorp/cli"
	"github.com/hashicorp/logutils"
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

	c := cli.NewCLI("hc-install", version.Version().String())
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"install": func() (cli.Command, error) {
			return &InstallCommand{
				Ui: ui,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
