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

var (
	cfgMgr    *cfg.Manager
	kernelMgr *kernel.Manager
	apiClient *api.Client
	subMgr    *subscription.Manager
)

func SetConfigManager(mgr *cfg.Manager)        { cfgMgr = mgr }
func SetKernelManager(mgr *kernel.Manager)      { kernelMgr = mgr }
func SetAPIClient(client *api.Client)           { apiClient = client }
func SetSubscriptionManager(mgr *subscription.Manager) { subMgr = mgr }

var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}
		return initBaseManagers()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ac, err := ensureMihomo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if ac == nil {
			fmt.Fprintln(os.Stderr, "Mihomo is not running. Start it with: mihomo-cli service start")
			os.Exit(1)
		}
		model := tui.NewModel(ac, kernelMgr, subMgr)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
	},
}

// initBaseManagers initializes config, kernel, and subscription managers.
// Lightweight — does not start mihomo.
func initBaseManagers() error {
	if cfgMgr != nil {
		return nil
	}

	cm, err := cfg.NewManager()
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}
	cfgMgr = cm

	c := cm.Config()
	kernelMgr = kernel.NewManager(cm.ConfigDir(), cm.MihomoDir(), c.Core.APIPort)

	subMgr = subscription.NewManager(cm)

	return nil
}

// ensureMihomo extracts the embedded kernel if needed, starts mihomo if not
// running, and returns an API client. Uses PID file to prevent duplicates.
func ensureMihomo() (*api.Client, error) {
	if kernelMgr == nil {
		return nil, fmt.Errorf("kernel manager not initialized")
	}

	if !kernelMgr.IsInstalled() {
		if err := kernelMgr.ExtractEmbedded(kernelMgr.BinPath()); err != nil {
			return nil, fmt.Errorf("kernel not installed: %w\n  Install manually: mihomo-cli kernel install --local <path>", err)
		}
	}

	if !kernelMgr.IsRunning() {
		if err := kernelMgr.Start(); err != nil {
			return nil, fmt.Errorf("start mihomo: %w", err)
		}
	}

	ac := kernelMgr.APIClient()
	if ac == nil {
		return nil, fmt.Errorf("mihomo API not reachable")
	}
	return ac, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
