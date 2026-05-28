package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/cfg"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
	"mihomo-cli/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	Run: func(cmd *cobra.Command, args []string) {
		cm, km, ac, sm, err := InitManagers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
		}
		if ac == nil {
			fmt.Println("Mihomo is not running. Start it with: mihomo-cli service start")
			os.Exit(1)
		}
		model := tui.NewModel(ac, km, sm)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		_ = cm
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

	if !km.IsInstalled() {
		if err := km.ExtractEmbedded(km.BinPath()); err != nil {
			fmt.Fprintf(os.Stderr, "error: extract kernel: %v\n", err)
			fmt.Fprintf(os.Stderr, "  Install manually: mihomo-cli kernel install --local <path>\n")
		}
	}

	if km.IsInstalled() && !km.IsRunning() {
		if err := km.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "error: start mihomo: %v\n", err)
		}
	}

	apiClient := km.APIClient()

	sm := subscription.NewManager(cm)

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
