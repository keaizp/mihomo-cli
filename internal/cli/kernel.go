package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kernelInstallURL string
var kernelInstallLocal string

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "Manage mihomo kernel binary",
}

var kernelInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Download or copy mihomo kernel binary",
	Long: `Install the mihomo kernel binary.

Without flags: downloads from GitHub (requires direct network access).

  mihomo-cli kernel install --url https://mirror.example.com/mihomo-linux-amd64.gz
  mihomo-cli kernel install --local ./mihomo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}

		// --local: copy from a local path
		if kernelInstallLocal != "" {
			fmt.Printf("Installing mihomo from %s...\n", kernelInstallLocal)
			if err := kernelMgr.InstallFrom(kernelInstallLocal); err != nil {
				return fmt.Errorf("install from local: %w", err)
			}
			fmt.Printf("mihomo installed to %s\n", kernelMgr.BinPath())
			return nil
		}

		// --url: download from custom URL
		if kernelInstallURL != "" {
			fmt.Printf("Downloading mihomo from %s...\n", kernelInstallURL)
			if err := kernelMgr.InstallFromURL(kernelInstallURL); err != nil {
				return fmt.Errorf("download from URL: %w", err)
			}
			fmt.Printf("mihomo installed to %s\n", kernelMgr.BinPath())
			return nil
		}

		// Default: download from GitHub
		fmt.Printf("Downloading mihomo from GitHub...\n")
		fmt.Printf("(If this fails due to network, use: mihomo-cli kernel install --local <path>)\n\n")
		if err := kernelMgr.Install(); err != nil {
			return fmt.Errorf("download from GitHub: %w\n\nTip: download manually and use --local:\n  mihomo-cli kernel install --local ./mihomo", err)
		}
		fmt.Printf("mihomo installed to %s\n", kernelMgr.BinPath())
		return nil
	},
}

var kernelPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show where the mihomo binary should be placed",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		fmt.Println(kernelMgr.BinPath())
		return nil
	},
}

func init() {
	kernelInstallCmd.Flags().StringVar(&kernelInstallURL, "url", "", "download from a custom URL (e.g. mirror)")
	kernelInstallCmd.Flags().StringVar(&kernelInstallLocal, "local", "", "copy from a local file path")

	kernelCmd.AddCommand(kernelInstallCmd)
	kernelCmd.AddCommand(kernelPathCmd)
	rootCmd.AddCommand(kernelCmd)
}
