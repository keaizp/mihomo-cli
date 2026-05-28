package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/subscription"
)

var subMgr *subscription.Manager

func SetSubscriptionManager(mgr *subscription.Manager) {
	subMgr = mgr
}

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscriptions",
}

var subAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a subscription",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.AddSubscription(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Subscription %q added\n", args[0])
		return nil
	},
}

var subRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.RemoveSubscription(args[0]); err != nil {
			return err
		}
		fmt.Printf("Subscription %q removed\n", args[0])
		return nil
	},
}

var subUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update subscriptions (all or by name)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if subMgr == nil {
			return fmt.Errorf("subscription manager not initialized")
		}
		if len(args) > 0 {
			if err := subMgr.UpdateSubscription(args[0]); err != nil {
				return err
			}
			fmt.Printf("Subscription %q updated\n", args[0])
		} else {
			errs := subMgr.UpdateAll()
			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, e)
				}
				return fmt.Errorf("%d subscription(s) failed to update", len(errs))
			}
			fmt.Println("All subscriptions updated")
		}
		return nil
	},
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		subs := cfgMgr.Config().Subscriptions
		if len(subs) == 0 {
			fmt.Println("No subscriptions configured")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL")
		for _, s := range subs {
			fmt.Fprintf(w, "%s\t%s\n", s.Name, s.URL)
		}
		w.Flush()
		return nil
	},
}

func init() {
	subCmd.AddCommand(subAddCmd)
	subCmd.AddCommand(subRemoveCmd)
	subCmd.AddCommand(subUpdateCmd)
	subCmd.AddCommand(subListCmd)
	rootCmd.AddCommand(subCmd)
}
