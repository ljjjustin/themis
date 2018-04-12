package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ljjjustin/themis/client"
	"github.com/spf13/cobra"
	texttable "github.com/syohex/go-texttable"
)

var (
	HostId   int
	IPMIHost string
	IPMIPort int
	Username string
	Password string
)

func NewFencerCommand() *cobra.Command {
	fencerCmd := &cobra.Command{
		Use:   "fencer",
		Short: "Host fencer related commands",
	}

	fencerCmd.AddCommand(newFencerListCommand())
	fencerCmd.AddCommand(newFencerGetCommand())
	fencerCmd.AddCommand(newFencerAddCommand())
	fencerCmd.AddCommand(newFencerDeleteCommand())

	return fencerCmd
}

func newFencerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all fencer",
		Run:   fencerListCommandFunc,
	}
	return cmd
}

func newFencerGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show a fencer details",
		Run:   fencerGetCommandFunc,
	}
	return cmd
}

func newFencerAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new fencer for host",
		Run:   fencerAddCommandFunc,
	}
	cmd.Flags().IntVarP(&HostId, "id", "I", 0, "host id")
	cmd.Flags().IntVarP(&IPMIPort, "port", "P", 623, "IPMI port")
	cmd.Flags().StringVarP(&IPMIHost, "host", "H", "", "IPMI Remote host name for LAN interface")
	cmd.Flags().StringVarP(&Username, "username", "u", "", "IPMI username")
	cmd.Flags().StringVarP(&Password, "password", "p", "", "IPMI password")

	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newFencerDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del <fencer id>",
		Short: "Delete a fencer",
		Run:   fencerDeleteCommandFunc,
	}
	return cmd
}

func displayFencers(fencers []client.Fencer) {
	table := &texttable.TextTable{}

	table.SetHeader("ID", "HostId", "Type",
		"Host", "Port", "Username")
	for _, f := range fencers {
		table.AddRow(
			fmt.Sprint(f.ID),
			fmt.Sprint(f.HostId),
			f.Type,
			f.Host,
			fmt.Sprint(f.Port),
			f.Username,
		)
	}

	fmt.Println(table.Draw())
}

func getFencerId(args []string) int {
	if len(args) != 1 {
		fmt.Println("ERROR: you must specify fencer id")
		os.Exit(-1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("ERROR: you must specify a valid id")
		os.Exit(-1)
	}
	return id
}

func fencerListCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)

	fencers, err := themis.ListFencers()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	displayFencers(fencers)
}

func fencerGetCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)

	fencer, err := themis.ShowFencer(getFencerId(args))
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	displayFencers([]client.Fencer{fencer})
}

func fencerAddCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)

	host, err := themis.ShowHost(HostId)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	req := &client.Fencer{
		HostId:   host.ID,
		Type:     "ipmi",
		Host:     IPMIHost,
		Port:     IPMIPort,
		Username: Username,
		Password: Password,
	}

	fencer, err := themis.AddFencer(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	displayFencers([]client.Fencer{fencer})
}

func fencerDeleteCommandFunc(cmd *cobra.Command, args []string) {
	themis := client.NewThemisClient(globalFlags.Url)
	err := themis.DeleteFencer(getFencerId(args))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
