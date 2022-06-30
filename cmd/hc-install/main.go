package main

import (
	"log"
	"os"

	"github.com/hashicorp/hc-install/internal/version"

	"github.com/hashicorp/logutils"
	"github.com/mitchellh/cli"
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

	c := cli.NewCLI("hc-install", version.ModuleVersion())
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
