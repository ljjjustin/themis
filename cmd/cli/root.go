package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type GlobalFlags struct {
	Debug bool
	Url   string
}

const (
	cliName        = "themisctl"
	cliDescription = "A simple command line client for themis."
)

var (
	rootCmd = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		SuggestFor: []string{"themisctl"},
	}
	globalFlags = GlobalFlags{}
)

func init() {

	rootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "enable client-side debug logging")
	rootCmd.PersistentFlags().StringVar(&globalFlags.Url, "url", "http://127.0.0.1:7878", "themis server URL")

	rootCmd.AddCommand(
		NewHostCommand(),
		NewFencerCommand(),
	)
}

func init() {
	cobra.EnablePrefixMatching = true
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}
}
