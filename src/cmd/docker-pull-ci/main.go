package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var version string
var builddate string
var versionString string
var debug bool
var quiet bool

func init() {
	// // Output to stdout instead of the default stderr
	// // Can be any io.Writer, see below for File example
	// logrus.SetOutput(os.Stderr)
	//
	// // Only log the warning severity or above.
	// logrus.SetLevel(logrus.InfoLevel)
	if builddate != "" {
		versionString = fmt.Sprintf("%s (%s)", version, builddate)
	} else {
		versionString = version
	}
	beforeAction()
}

func beforeAction() {
	// Only log the warning severity or above.
	logrus.Debugf("Debug: %t", debug)
	logrus.Debugf("Quiet: %t", quiet)
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if quiet {
		logrus.SetLevel(logrus.WarnLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetOutput(os.Stderr)
}

func main() {

	debugFlag := cli.BoolFlag{
		Name:        "debug, d",
		Usage:       "Enable Debugging",
		Destination: &debug,
	}
	quietFlag := cli.BoolFlag{
		Name:        "quiet, q",
		Usage:       "Only output warnings or fatals",
		Destination: &quiet,
	}
	app := cli.NewApp()

	app.Name = "docker-pull-ci"
	app.Usage = "Docker Pull Output for CI"
	app.Version = versionString
	app.Flags = []cli.Flag{debugFlag, quietFlag}
	app.Action = parseCommand

	app.Run(os.Args)
}
