package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage mihomo-cli configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		cfg := cfgMgr.Config()
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("HTTP Port: %d\n", cfg.Core.HTTPPort)
		fmt.Printf("SOCKS Port: %d\n", cfg.Core.SOCKSPort)
		fmt.Printf("API Port: %d\n", cfg.Core.APIPort)
		fmt.Printf("Subscriptions: %d\n", len(cfg.Subscriptions))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit config with $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
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
	Short: "Reload mihomo configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ac, err := ensureMihomo()
		if err != nil {
			return err
		}
		if err := ac.ReloadConfig(); err != nil {
			return fmt.Errorf("reload config: %w", err)
		}
		fmt.Println("Configuration reloaded")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configReloadCmd)
	rootCmd.AddCommand(configCmd)
}
