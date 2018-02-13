package cmd

import (
	"fmt"
	"os"

	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/cobra"
)

const (
	cliName = "themis"
	cliDesc = "Command line for themis."
)

// Command line configurations
var (
	configFile string
)

var rootCmd = &cobra.Command{
	Use:        cliName,
	Short:      cliDesc,
	SuggestFor: []string{"themis"},
}

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "cmd")

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}
}

func init() {
	rootCmd.AddCommand(
		NewMonitorCommand(),
		NewDbsyncCommand(),
	)
}

func init() {
	cobra.EnablePrefixMatching = true
}
