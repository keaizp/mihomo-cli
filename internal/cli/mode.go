package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get or set proxy mode",
}

var modeSetCmd = &cobra.Command{
	Use:   "set <rule|global|direct|script>",
	Short: "Set proxy mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.SetMode(args[0]); err != nil {
			return err
		}
		fmt.Printf("Mode set to: %s\n", args[0])
		return nil
	},
}

var modeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current proxy mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
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
