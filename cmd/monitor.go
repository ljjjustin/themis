package cmd

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/monitor"
	"github.com/spf13/cobra"
)

func NewMonitorCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "monitor",
		Short: "Perform monitoring and fence operation.",
		Run:   monitorMain,
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to toml config file.")

	return &cmd
}

func monitorMain(cmd *cobra.Command, args []string) {

	// load configurations
	themisCfg := config.NewConfig(configFile)

	// init log configurations
	themisCfg.SetupLogging()

	plog.Println("Starting monitor server...")
	monitor := monitor.NewThemisMonitor(themisCfg)
	monitor.Start()
}
