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
	Short: "管理代理节点",
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
	Short: "列出代理组和节点",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("获取代理列表失败: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "代理组\t当前节点\t可选节点")
		for name, p := range proxies.Proxies {
			if p.All != nil {
				nodes := ""
				for i, n := range p.All {
					marker := ""
					if n == p.Now {
						marker = "● "
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
	Use:   "set <代理组> <节点名>",
	Short: "切换节点",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}
		if err := apiClient.SwitchProxy(args[0], args[1]); err != nil {
			return fmt.Errorf("切换失败: %w", err)
		}
		fmt.Printf("✓ 已切换 [%s] → %s\n", args[0], args[1])
		return nil
	},
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [节点名]",
	Short: "测试节点延迟",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}

		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("获取代理列表失败: %w", err)
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
		fmt.Fprintln(w, "节点\t延迟")
		for _, r := range all {
			fmt.Fprintf(w, "%s\t%dms\n", r.name, r.delay)
		}
		w.Flush()
		return nil
	},
}

var proxyInfoCmd = &cobra.Command{
	Use:   "info <节点名>",
	Short: "查看节点详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API 不可用，mihomo 是否在运行？")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("获取代理列表失败: %w", err)
		}
		p, ok := proxies.Proxies[args[0]]
		if !ok {
			return fmt.Errorf("节点 %q 不存在", args[0])
		}
		fmt.Printf("名称: %s\n类型: %s\n", p.Name, p.Type)
		if len(p.History) > 0 {
			fmt.Printf("最近延迟: %dms\n", p.History[len(p.History)-1].Delay)
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
