// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/mitchellh/cli"
)

var version = "0.0.0"

func main() {
	c := cli.NewCLI("terraform", version)
	c.HelpWriter = os.Stdout
	c.Args = os.Args[1:]

	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}

	c.Commands = map[string]cli.CommandFactory{
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				Ui:      ui,
				Version: version,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

type VersionOutput struct {
	Version            string            `json:"terraform_version"`
	Platform           string            `json:"platform"`
	ProviderSelections map[string]string `json:"provider_selections"`
	Outdated           bool              `json:"terraform_outdated"`
}

type VersionCommand struct {
	Ui      cli.Ui
	Version string

	jsonOutput bool
}

func (c *VersionCommand) flags() *flag.FlagSet {
	fs := defaultFlagSet("version")

	fs.BoolVar(&c.jsonOutput, "json", false, "output the version information as a JSON object")

	fs.Usage = func() { c.Ui.Error(c.Help()) }

	return fs
}

func (c *VersionCommand) Run(args []string) int {
	f := c.flags()
	if err := f.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing command-line flags: %s", err))
		return 1
	}

	if c.jsonOutput {
		output := VersionOutput{
			Version:  c.Version,
			Platform: runtime.GOOS + "_" + runtime.GOARCH,
		}

		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			c.Ui.Error(fmt.Sprintf("\nError marshalling JSON: %s", err))
			return 1
		}
		c.Ui.Output(string(jsonOutput))
	} else {
		c.Ui.Output(fmt.Sprintf("Terraform v%s", c.Version))
	}

	return 0
}

func defaultFlagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.Usage = func() {}
	return f
}

func (c *VersionCommand) Help() string {
	helpText := `
Usage: terraform version [-json]
` + c.Synopsis()

	return strings.TrimSpace(helpText)
}

func (c *VersionCommand) Synopsis() string {
	return "Displays the version"
}
