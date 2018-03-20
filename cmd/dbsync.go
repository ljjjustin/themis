package cmd

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/storage"

	"github.com/spf13/cobra"
)

func NewDbsyncCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "dbsync",
		Short: "Perform Database Model synchronize.",
		Run:   dbsyncMain,
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to toml config file.")

	return &cmd
}

func dbsyncMain(cmd *cobra.Command, args []string) {

	plog.Println("Parse config and loading config file...")

	// load configurations
	themisCfg := config.NewConfig(configFile)

	// init log configurations
	themisCfg.SetupLogging()

	storage.DbSync(&themisCfg.Storage)
}
