package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ljjjustin/themis/client"
	"github.com/spf13/cobra"
	texttable "github.com/syohex/go-texttable"
)

// NewHostCommand returns the cobra command for "Host".
func NewHostCommand() *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "host",
		Short: "Host related commands",
	}

	hostCmd.AddCommand(newHostAddCommand())
	hostCmd.AddCommand(newHostDeleteCommand())
	hostCmd.AddCommand(newHostGetCommand())
	hostCmd.AddCommand(newHostListCommand())
	hostCmd.AddCommand(newHostEnableCommand())
	hostCmd.AddCommand(newHostDisableCommand())

	return hostCmd
}

func newHostAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <hostname>",
		Short: "Adds a new host",
		Run:   hostAddCommandFunc,
	}
	return cmd
}

func newHostDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del <host id>",
		Short: "Delete a host",
		Run:   hostDeleteCommandFunc,
	}
	return cmd
}

func newHostGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <host id>",
		Short: "Get a host information",
		Run:   hostGetCommandFunc,
	}
	return cmd
}
func newHostListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all host information",
		Run:   hostListCommandFunc,
	}
	return cmd
}

func newHostEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <host id>",
		Short: "Enable a host",
		Run:   hostEnableCommandFunc,
	}
	return cmd
}

func newHostDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <host id>",
		Short: "Disable a host",
		Run:   hostDisableCommandFunc,
	}
	return cmd
}

func displayHosts(hosts []client.Host) {
	table := &texttable.TextTable{}

	table.SetHeader("ID", "Name", "Status", "Disabled", "UpdatedAt")
	for _, h := range hosts {
		table.AddRow(
			fmt.Sprint(h.ID),
			h.Name, h.Status,
			fmt.Sprint(h.Disabled),
			h.UpdatedAt.Format(time.RFC3339),
		)
	}

	fmt.Println(table.Draw())
}

func hostListCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)

	hosts, err := themis.ListHosts()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	displayHosts(hosts)
}

func getHostId(args []string) int {
	if len(args) != 1 {
		fmt.Println("ERROR: you must specify host id")
		os.Exit(-1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("ERROR: you must specify a valid id")
		os.Exit(-1)
	}
	return id
}

func hostGetCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)
	host, err := themis.ShowHost(getHostId(args))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	displayHosts([]client.Host{host})
}

func hostAddCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("ERROR: you must specify hostname")
		os.Exit(-1)
	}
	req := &client.Host{Name: args[0]}

	themis := client.NewThemisClient(globalFlags.Url)
	host, err := themis.AddHost(req)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	displayHosts([]client.Host{host})
}

func hostDeleteCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)
	err := themis.DeleteHost(getHostId(args))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func hostEnableCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)
	host, err := themis.EnableHost(getHostId(args))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	displayHosts([]client.Host{host})
}

func hostDisableCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)
	host, err := themis.DisableHost(getHostId(args))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	displayHosts([]client.Host{host})
}
