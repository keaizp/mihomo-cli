package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "管理 mihomo 服务",
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "启动 mihomo",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		if kernelMgr.IsRunning() {
			fmt.Println("✓ 已运行中")
			return nil
		}
		if !kernelMgr.IsInstalled() {
			if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
				return fmt.Errorf("内核未安装，请先运行 mihomo-cli 自动部署")
			}
		}
		if err := kernelMgr.Start(); err != nil {
			return fmt.Errorf("启动失败: %w", err)
		}
		fmt.Println("✓ 已启动")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止 mihomo",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		if err := kernelMgr.Stop(); err != nil {
			return fmt.Errorf("停止失败: %w", err)
		}
		fmt.Println("✓ 已停止")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启 mihomo",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		if !kernelMgr.IsInstalled() {
			if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
				return fmt.Errorf("内核未安装")
			}
		}
		if err := kernelMgr.Restart(); err != nil {
			return fmt.Errorf("重启失败: %w", err)
		}
		fmt.Println("✓ 已重启")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看运行状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		fmt.Println(kernelMgr.Status())
		return nil
	},
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看日志",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		lines, err := kernelMgr.ReadLogs(50)
		if err != nil {
			return fmt.Errorf("读取日志失败: %w", err)
		}
		if len(lines) == 0 {
			fmt.Println("(暂无日志)")
			return nil
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	},
}

var servicePrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "准备内核和配置（systemd ExecStartPre 使用）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		if !kernelMgr.IsInstalled() {
			if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
				return fmt.Errorf("内核未安装: %w", err)
			}
		}
		if subMgr != nil {
			if err := subMgr.MergeAndGenerate(); err != nil {
				return fmt.Errorf("生成配置失败: %w", err)
			}
		}
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	serviceCmd.AddCommand(servicePrepareCmd)
	rootCmd.AddCommand(serviceCmd)
}
