package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/ljjjustin/themis/client"
	"github.com/spf13/cobra"
	texttable "github.com/syohex/go-texttable"
)

// NewLeaderCommand returns the cobra command for "Leader".
func NewLeaderCommand() *cobra.Command {
	leaderCmd := &cobra.Command{
		Use:   "leader",
		Short: "Show leader",
		Run:   showLeaderCommandFunc,
	}

	return leaderCmd
}

func showLeaderCommandFunc(cmd *cobra.Command, args []string) {

	themis := client.NewThemisClient(globalFlags.Url)

	leader, err := themis.GetLeader()
	if err != nil {
		fmt.Println("server err :", err)
		os.Exit(-1)
	}
	displayLeader([]client.ElectionRecord{leader})
}

func displayLeader(leader []client.ElectionRecord) {
	table := &texttable.TextTable{}

	table.SetHeader("LeaderName", "LastUpdate")
	for _, l := range leader {
		table.AddRow(
			l.LeaderName,
			l.LastUpdate.Format(time.RFC3339),
		)
	}

	fmt.Println(table.Draw())
}