package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage proxy nodes",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ac, err := ensureMihomo()
		if err != nil {
			return err
		}
		apiClient = ac
		return nil
	},
}

var proxyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List proxy groups and nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "GROUP\tCURRENT\tNODES")
		for name, p := range proxies.Proxies {
			if p.All != nil {
				nodes := ""
				for i, n := range p.All {
					marker := ""
					if n == p.Now {
						marker = "*"
					}
					if i > 0 {
						nodes += ", "
					}
					nodes += marker + n
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", name, p.Now, nodes)
			}
		}
		w.Flush()
		return nil
	},
}

var proxySetCmd = &cobra.Command{
	Use:   "set <group> <node>",
	Short: "Switch proxy node in a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		if err := apiClient.SwitchProxy(args[0], args[1]); err != nil {
			return fmt.Errorf("switch proxy: %w", err)
		}
		fmt.Printf("Switched [%s] → %s\n", args[0], args[1])
		return nil
	},
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [node]",
	Short: "Test proxy latency",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}

		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}

		type result struct {
			name  string
			delay int
			err   error
		}

		results := make(chan result, 50)
		count := 0

		for name, p := range proxies.Proxies {
			if p.All != nil {
				continue
			}
			if len(args) > 0 && name != args[0] {
				continue
			}
			count++
			go func(n string) {
				d, err := apiClient.TestDelay(n, 5*time.Second)
				results <- result{name: n, delay: d, err: err}
			}(name)
		}

		var all []result
		for i := 0; i < count; i++ {
			r := <-results
			if r.err == nil {
				all = append(all, r)
			}
		}
		sort.Slice(all, func(i, j int) bool { return all[i].delay < all[j].delay })

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NODE\tDELAY")
		for _, r := range all {
			fmt.Fprintf(w, "%s\t%dms\n", r.name, r.delay)
		}
		w.Flush()
		return nil
	},
}

var proxyInfoCmd = &cobra.Command{
	Use:   "info <node>",
	Short: "Show proxy node details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}
		p, ok := proxies.Proxies[args[0]]
		if !ok {
			return fmt.Errorf("node %q not found", args[0])
		}
		fmt.Printf("Name: %s\nType: %s\n", p.Name, p.Type)
		if len(p.History) > 0 {
			fmt.Printf("Last delay: %dms\n", p.History[len(p.History)-1].Delay)
		}
		return nil
	},
}

func init() {
	proxyCmd.AddCommand(proxyListCmd)
	proxyCmd.AddCommand(proxySetCmd)
	proxyCmd.AddCommand(proxyTestCmd)
	proxyCmd.AddCommand(proxyInfoCmd)
	rootCmd.AddCommand(proxyCmd)
}
