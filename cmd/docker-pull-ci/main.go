package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint: gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use:   "docker-pull-ci",
	Short: "Command to parse output of docker pull and docker push for readability on long push/pulls",
	Run:   parseCommand,
	Args:  cobra.NoArgs,
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(beforeAction)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet output")
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindEnv("quiet", "QUIET")
}

func beforeAction() {
	switch {
	case viper.GetBool("debug"):
		logrus.SetLevel(logrus.DebugLevel)
	case viper.GetBool("quiet"):
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetOutput(os.Stderr)
}

func main() {
	_ = rootCmd.Execute()
}
