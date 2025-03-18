package main

import (
	"os"

	"github.com/dosquad/go-cliversion"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "docker-pull-ci",
	Short:   "Command to parse output of docker pull and docker push for readability on long push/pulls",
	Run:     parseCommand,
	Args:    cobra.NoArgs,
	Version: cliversion.Get().VersionString(),
}

func init() {
	cobra.OnInitialize(beforeAction)
}

func beforeAction() {
	logrus.SetOutput(os.Stderr)
}

func main() {
	_ = rootCmd.Execute()
}
