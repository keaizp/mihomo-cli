package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/cfg"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
)

var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TUI mode not yet implemented")
		os.Exit(0)
	},
}

// InitManagers creates and wires all managers from config.
func InitManagers() (*cfg.Manager, *kernel.Manager, *api.Client, *subscription.Manager, error) {
	cm, err := cfg.NewManager()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init config: %w", err)
	}

	c := cm.Config()
	km := kernel.NewManager(cm.ConfigDir(), cm.MihomoDir(), c.Core.APIPort)

	// Auto-install kernel
	if !km.IsInstalled() {
		fmt.Println("Downloading mihomo kernel...")
		if err := km.Install(); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("install mihomo: %w", err)
		}
		fmt.Println("Mihomo kernel installed")
	}

	// Try to start if not running
	if !km.IsRunning() {
		fmt.Println("Starting mihomo...")
		if err := km.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not start mihomo: %v\n", err)
		}
	}

	apiClient := km.APIClient()

	sm := subscription.NewManager(cm)

	// Wire into CLI globals
	SetConfigManager(cm)
	SetKernelManager(km)
	if apiClient != nil {
		SetAPIClient(apiClient)
	}
	SetSubscriptionManager(sm)

	return cm, km, apiClient, sm, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
