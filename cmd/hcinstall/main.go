package main

// import (
// 	"context"
// 	"flag"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"strings"

// 	"github.com/hashicorp/logutils"
// 	"github.com/mitchellh/cli"

// 	"github.com/hashicorp/hcinstall"
// )

// // TODO: add versioning to this?
// const userAgentAppend = "hcinstall-cli"

// func main() {
// 	filter := &logutils.LevelFilter{
// 		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
// 		MinLevel: logutils.LogLevel("WARN"),
// 		Writer:   os.Stderr,
// 	}
// 	log.SetOutput(filter)

// 	ui := &cli.ColoredUi{
// 		ErrorColor: cli.UiColorRed,
// 		WarnColor:  cli.UiColorYellow,
// 		Ui: &cli.BasicUi{
// 			Reader:      os.Stdin,
// 			Writer:      os.Stdout,
// 			ErrorWriter: os.Stderr,
// 		},
// 	}

// 	exitStatus := run(ui, os.Args[1:])

// 	os.Exit(exitStatus)
// }

// func help() string {
// 	return `Usage: hcinstall [--dir=DIR] VERSION-OR-REF

//   Downloads, verifies, and installs a official releases of a binary
//   from releases.hashicorp.com or downloads, compiles, and installs a version of
//   the the binary from the GitHub repository.

//   To download an official release, pass "latest" or a valid semantic versioning
//   version string.

//   To download and compile a version of the binary from the GitHub
//   repository pass a ref in the form "refs/...", some examples are shown below.

//   If a binary is successfully installed, its path will be printed to stdout.

//   Unless --dir is given, the default system temporary directory will be used.

// Options:
//   --dir          Directory into which to install the terraform binary. The
//                  directory must exist.

// Examples:
//   hcinstall terraform 0.12.28
//   hcinstall consul latest
//   hcinstall terraform 0.13.0-beta3
//   hcinstall --dir=/home/kmoe/bin 0.12.28
//   hcinstall refs/heads/master
//   hcinstall refs/tags/v0.12.29
//   hcinstall refs/pull/25633/head
// `
// }

// func run(ui cli.Ui, args []string) int {
// 	ctx := context.Background()

// 	args = os.Args[1:]
// 	flags := flag.NewFlagSet("", flag.ExitOnError)
// 	var tfDir string
// 	flags.StringVar(&tfDir, "dir", "", "Local directory into which to install terraform")

// 	err := flags.Parse(args)
// 	if err != nil {
// 		ui.Error(err.Error())
// 		return 1
// 	}

// 	if flags.NArg() != 1 {
// 		ui.Error("Please specify VERSION-OR-REF")
// 		ui.Output(help())
// 		return 127
// 	}

// 	tfVersion := flags.Args()[0]

// 	if tfDir == "" {
// 		tfDir, err = ioutil.TempDir("", "hcinstall")
// 		if err != nil {
// 			ui.Error(err.Error())
// 			return 1
// 		}
// 	}

// 	var findArgs []hcinstall.ExecPathFinder

// 	switch {
// 	case tfVersion == "latest":
// 		finder := hcinstall.LatestVersion(tfDir, false)
// 		finder.UserAgent = userAgentAppend
// 		findArgs = append(findArgs, finder)
// 	case strings.HasPrefix(tfVersion, "refs/"):
// 		findArgs = append(findArgs, gitref.Install(tfVersion, "", tfDir))
// 	default:
// 		if strings.HasPrefix(tfVersion, "v") {
// 			tfVersion = tfVersion[1:]
// 		}
// 		finder := hcinstall.ExactVersion(tfVersion, tfDir)
// 		finder.UserAgent = userAgentAppend
// 		findArgs = append(findArgs, finder)
// 	}

// 	tfPath, err := hcinstall.Find(ctx, findArgs...)
// 	if err != nil {
// 		ui.Error(err.Error())
// 		return 1
// 	}

// 	ui.Output(tfPath)
// 	return 0
// }
