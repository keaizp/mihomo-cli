package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "查看当前配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		cfg := cfgMgr.Config()
		fmt.Printf("模式: %s\n", cfg.Mode)
		fmt.Printf("HTTP 端口: %d\n", cfg.Core.HTTPPort)
		fmt.Printf("SOCKS 端口: %d\n", cfg.Core.SOCKSPort)
		fmt.Printf("API 端口: %d\n", cfg.Core.APIPort)
		fmt.Printf("订阅数: %d\n", len(cfg.Subscriptions))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "编辑配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		configPath := cfgMgr.ConfigDir() + "/config.yaml"
		c := exec.Command(editor, configPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var configReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "重载配置（无需重启）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ac, err := ensureMihomo()
		if err != nil {
			return err
		}
		if err := ac.ReloadConfig(); err != nil {
			return fmt.Errorf("重载失败: %w", err)
		}
		fmt.Println("✓ 配置已重载")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configReloadCmd)
	rootCmd.AddCommand(configCmd)
}
