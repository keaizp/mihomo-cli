package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "Manage active connections",
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		conns, err := apiClient.GetConnections()
		if err != nil {
			return fmt.Errorf("get connections: %w", err)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tHOST\tNETWORK\tRULE\tUPLOAD\tDOWNLOAD")
		for _, c := range conns.Connections {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\n",
				c.ID, c.Metadata.Host, c.Metadata.Network, c.Rule, c.Upload, c.Download)
		}
		w.Flush()
		return nil
	},
}

var connCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		if err := apiClient.CloseConnection(args[0]); err != nil {
			return fmt.Errorf("close connection: %w", err)
		}
		fmt.Printf("Connection %s closed\n", args[0])
		return nil
	},
}

func init() {
	connCmd.AddCommand(connListCmd)
	connCmd.AddCommand(connCloseCmd)
	rootCmd.AddCommand(connCmd)
}
