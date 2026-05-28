package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage mihomo kernel service",
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if kernelMgr.IsRunning() {
			fmt.Println("mihomo is already running")
			return nil
		}
		if !kernelMgr.IsInstalled() {
			if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
				return fmt.Errorf("kernel not installed: install via 'mihomo-cli kernel install'")
			}
		}
		if err := kernelMgr.Start(); err != nil {
			return fmt.Errorf("start mihomo: %w", err)
		}
		fmt.Println("mihomo started")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if err := kernelMgr.Stop(); err != nil {
			return fmt.Errorf("stop mihomo: %w", err)
		}
		fmt.Println("mihomo stopped")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if !kernelMgr.IsInstalled() {
			if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
				return fmt.Errorf("kernel not installed: install via 'mihomo-cli kernel install'")
			}
		}
		if err := kernelMgr.Restart(); err != nil {
			return fmt.Errorf("restart mihomo: %w", err)
		}
		fmt.Println("mihomo restarted")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mihomo kernel status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		fmt.Printf("mihomo: %s\n", kernelMgr.Status())
		return nil
	},
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show mihomo kernel logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		lines, err := kernelMgr.ReadLogs(50)
		if err != nil {
			return fmt.Errorf("read logs: %w", err)
		}
		if len(lines) == 0 {
			fmt.Println("(no logs)")
			return nil
		}
		for _, line := range lines {
			fmt.Println(line)
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
	rootCmd.AddCommand(serviceCmd)
}
