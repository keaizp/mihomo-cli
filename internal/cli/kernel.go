package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kernelInstallURL string
var kernelInstallLocal string

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "管理 mihomo 内核（正常使用不需要关心）",
}

var kernelInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装或替换 mihomo 内核",
	Long: `从本地文件或 URL 安装 mihomo 内核。正常使用不需要此命令，内核已嵌入。

  mihomo-cli kernel install --local ./mihomo-linux-amd64.gz
  mihomo-cli kernel install --url https://mirror.example.com/mihomo.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}

		if kernelInstallLocal != "" {
			if err := kernelMgr.InstallFrom(kernelInstallLocal); err != nil {
				return fmt.Errorf("安装本地内核失败: %w", err)
			}
			fmt.Printf("✓ 内核已安装: %s\n", kernelMgr.BinPath())
			return nil
		}

		if kernelInstallURL != "" {
			if err := kernelMgr.InstallFromURL(kernelInstallURL); err != nil {
				return fmt.Errorf("下载内核失败: %w", err)
			}
			fmt.Printf("✓ 内核已安装: %s\n", kernelMgr.BinPath())
			return nil
		}

		// Default: download from GitHub
		if err := kernelMgr.Install(); err != nil {
			return fmt.Errorf("下载失败: %w\n\n提示：手动下载后使用 --local 安装", err)
		}
		fmt.Printf("✓ 内核已安装: %s\n", kernelMgr.BinPath())
		return nil
	},
}

var kernelPathCmd = &cobra.Command{
	Use:   "path",
	Short: "查看内核路径",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("内核管理器未初始化")
		}
		fmt.Println(kernelMgr.BinPath())
		return nil
	},
}

func init() {
	kernelInstallCmd.Flags().StringVar(&kernelInstallURL, "url", "", "从指定 URL 下载")
	kernelInstallCmd.Flags().StringVar(&kernelInstallLocal, "local", "", "从本地文件安装")

	kernelCmd.AddCommand(kernelInstallCmd)
	kernelCmd.AddCommand(kernelPathCmd)
	rootCmd.AddCommand(kernelCmd)
}
