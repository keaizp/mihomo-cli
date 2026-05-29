package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "管理活跃连接",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ac, err := ensureMihomo()
		if err != nil {
			return err
		}
		apiClient = ac
		return nil
	},
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看活跃连接",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}
		conns, err := apiClient.GetConnections()
		if err != nil {
			return fmt.Errorf("获取连接失败: %w", err)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t主机\t协议\t规则\t上行\t下行")
		for _, c := range conns.Connections {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\n",
				c.ID, c.Metadata.Host, c.Metadata.Network, c.Rule, c.Upload, c.Download)
		}
		w.Flush()
		return nil
	},
}

var connCloseCmd = &cobra.Command{
	Use:   "close <ID>",
	Short: "关闭连接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}
		if err := apiClient.CloseConnection(args[0]); err != nil {
			return fmt.Errorf("关闭连接失败: %w", err)
		}
		fmt.Printf("✓ 已关闭连接 %s\n", args[0])
		return nil
	},
}

func init() {
	connCmd.AddCommand(connListCmd)
	connCmd.AddCommand(connCloseCmd)
	rootCmd.AddCommand(connCmd)
}
