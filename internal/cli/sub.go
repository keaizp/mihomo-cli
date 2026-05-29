package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "管理订阅",
}

var subAddCmd = &cobra.Command{
	Use:   "add <名称> <URL>",
	Short: "添加订阅",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		if err := cfgMgr.AddSubscription(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("✓ 已添加订阅: %s\n", args[0])
		return nil
	},
}

var subRemoveCmd = &cobra.Command{
	Use:   "remove <名称>",
	Short: "删除订阅",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		if err := cfgMgr.RemoveSubscription(args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ 已删除订阅: %s\n", args[0])
		return nil
	},
}

var subUpdateCmd = &cobra.Command{
	Use:   "update [名称]",
	Short: "更新订阅（不指定名称则更新全部）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if subMgr == nil {
			return fmt.Errorf("订阅管理器未初始化")
		}
		if len(args) > 0 {
			if err := subMgr.UpdateSubscription(args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ 已更新: %s\n", args[0])
		} else {
			errs := subMgr.UpdateAll()
			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, e)
				}
				return fmt.Errorf("%d 个订阅更新失败", len(errs))
			}
			fmt.Println("✓ 订阅已全部更新")
		}
		return nil
	},
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有订阅",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		subs := cfgMgr.Config().Subscriptions
		if len(subs) == 0 {
			fmt.Println("暂无订阅，使用 sub add 添加")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "名称\tURL")
		for _, s := range subs {
			fmt.Fprintf(w, "%s\t%s\n", s.Name, s.URL)
		}
		w.Flush()
		return nil
	},
}

func init() {
	subCmd.AddCommand(subAddCmd)
	subCmd.AddCommand(subRemoveCmd)
	subCmd.AddCommand(subUpdateCmd)
	subCmd.AddCommand(subListCmd)
	rootCmd.AddCommand(subCmd)
}
