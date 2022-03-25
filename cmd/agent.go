package cmd

import (
	"themis/config"
	"themis/monitor"
	"github.com/spf13/cobra"
)

func NewAgentCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "agent",
		Short: "monitoring serf agents.",
		Run:   agentMain,
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to toml config file.")

	return &cmd
}

func agentMain(cmd *cobra.Command, args []string) {

	// load configurations
	themisCfg := config.NewConfig(configFile)

	// init log configurations
	themisCfg.SetupLogging()

	agent := monitor.NewThemisAgent(themisCfg)
	agent.Start()
}
