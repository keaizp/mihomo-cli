package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "查看或切换代理模式",
}

var modeSetCmd = &cobra.Command{
	Use:   "set <rule|global|direct|script>",
	Short: "切换代理模式",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		if err := cfgMgr.SetMode(args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ 已切换为 %s 模式\n", args[0])
		return nil
	},
}

var modeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "查看当前模式",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		fmt.Println(cfgMgr.Config().Mode)
		return nil
	},
}

func init() {
	modeCmd.AddCommand(modeSetCmd)
	modeCmd.AddCommand(modeShowCmd)
	rootCmd.AddCommand(modeCmd)
}
