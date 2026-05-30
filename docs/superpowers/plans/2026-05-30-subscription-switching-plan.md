# Subscription Switching Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add active-subscription switching (single-selection), fix keybinding conflicts, and complete TUI subscription tab handlers.

**Architecture:** Add `ActiveSubscription` field to `AppConfig`, filter `MergeAndGenerate` by active sub, add `sub switch`/`sub edit` CLI commands, fix TUI keybinding conflict (`a`→`n`), and implement missing TUI sub tab handlers.

**Tech Stack:** Go, cobra CLI, bubbletea TUI, lipgloss styling

---

### Task 1: Add ActiveSubscription field and methods to cfg.Manager

**Files:**
- Modify: `internal/cfg/manager.go`

- [ ] **Step 1: Add ActiveSubscription to AppConfig struct**

```go
// AppConfig is the top-level mihomo-cli configuration.
type AppConfig struct {
	Core              CoreConfig       `yaml:"core"`
	ActiveSubscription string          `yaml:"active_subscription"`
	Subscriptions     []Subscription   `yaml:"subscriptions"`
	Mode              string           `yaml:"mode"` // rule, global, direct, script
	UserRules         []string         `yaml:"user_rules"`
	UserProxies       []map[string]any `yaml:"user_proxies"`
}
```

- [ ] **Step 2: Add UpdateSubscription method after RemoveSubscription**

```go
// UpdateSubscription updates a subscription's URL by name.
func (m *Manager) UpdateSubscription(name, url string) error {
	for i, s := range m.config.Subscriptions {
		if s.Name == name {
			m.config.Subscriptions[i].URL = url
			return m.Save()
		}
	}
	return fmt.Errorf("subscription %q not found", name)
}
```

- [ ] **Step 3: Add SetActiveSubscription method after UpdateSubscription**

```go
// SetActiveSubscription sets the active subscription and saves.
// Pass empty string to use all subscriptions.
func (m *Manager) SetActiveSubscription(name string) error {
	if name != "" {
		found := false
		for _, s := range m.config.Subscriptions {
			if s.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("subscription %q not found", name)
		}
	}
	m.config.ActiveSubscription = name
	return m.Save()
}
```

- [ ] **Step 4: Build and verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/cfg/manager.go
git commit -m "feat: add ActiveSubscription field and methods to cfg.Manager"
```

---

### Task 2: Filter MergeAndGenerate by active subscription

**Files:**
- Modify: `internal/subscription/manager.go:121-164`

- [ ] **Step 1: Update MergeAndGenerate to filter by active sub**

Replace the proxy-merging loop in `MergeAndGenerate` (lines 124-139):

```go
// MergeAndGenerate merges all subscription configs with user overrides and
// writes the final mihomo config file.
func (m *Manager) MergeAndGenerate() error {
	appCfg := m.cfg.Config()
	allProxies := make([]map[string]any, 0)

	profilesDir := filepath.Join(m.cfg.ConfigDir(), "profiles")
	for _, sub := range appCfg.Subscriptions {
		// If an active subscription is set, only merge that one.
		if appCfg.ActiveSubscription != "" && sub.Name != appCfg.ActiveSubscription {
			continue
		}
		profilePath := filepath.Join(profilesDir, sub.Name+".yaml")
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue
		}
		var sc SubscriptionConfig
		if err := yaml.Unmarshal(data, &sc); err != nil {
			continue
		}
		allProxies = append(allProxies, sc.Proxies...)
	}

	allProxies = append(allProxies, appCfg.UserProxies...)
	// ... rest unchanged
```

- [ ] **Step 2: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/subscription/manager.go
git commit -m "feat: filter MergeAndGenerate by active subscription"
```

---

### Task 3: Add sub switch and sub edit CLI commands

**Files:**
- Modify: `internal/cli/sub.go`

- [ ] **Step 1: Add subSwitchCmd after subListCmd**

```go
var subSwitchCmd = &cobra.Command{
	Use:   "switch <名称>",
	Short: "切换激活订阅（留空则使用全部）",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if err := cfgMgr.SetActiveSubscription(name); err != nil {
			return err
		}
		if name == "" {
			fmt.Println("✓ 已切换为使用全部订阅")
		} else {
			fmt.Printf("✓ 已切换激活订阅: %s\n", name)
		}
		// Regenerate mihomo config and reload
		if subMgr != nil {
			subMgr.MergeAndGenerate()
		}
		if kernelMgr != nil && kernelMgr.IsRunning() {
			if ac := kernelMgr.APIClient(); ac != nil {
				ac.ReloadConfig()
			}
		}
		return nil
	},
}

var subEditCmd = &cobra.Command{
	Use:   "edit <名称> <新URL>",
	Short: "编辑订阅 URL",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("配置管理器未初始化")
		}
		if err := cfgMgr.UpdateSubscription(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("✓ 已更新订阅: %s\n", args[0])
		return nil
	},
}
```

- [ ] **Step 2: Register new commands in init()**

Add to the `init()` function after the existing `subCmd.AddCommand` calls:

```go
	subCmd.AddCommand(subSwitchCmd)
	subCmd.AddCommand(subEditCmd)
```

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/cli/sub.go
git commit -m "feat: add sub switch and sub edit CLI commands"
```

---

### Task 4: Fix TUI keybinding conflicts

**Files:**
- Modify: `internal/tui/model.go:174-201`

- [ ] **Step 1: Change SubAdd key from "a" to "n", remove SubToggle**

In the `Keys` variable, replace:

```go
	SubAdd:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "添加订阅")),
```

with:

```go
	SubAdd:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "添加订阅")),
```

And remove the `SubToggle` line entirely:

```go
	// Remove this line:
	SubToggle: key.NewBinding(key.WithKeys("space"), key.WithHelp("空格", "启用/停用")),
```

Also remove `SubToggle` from the `KeyMap` struct:

```go
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	TabPrev   key.Binding
	TabNext   key.Binding
	Test      key.Binding
	TestAll   key.Binding
	Quit      key.Binding
	Reload    key.Binding
	Update    key.Binding
	Mode      key.Binding
	Search    key.Binding
	Collapse  key.Binding
	Close     key.Binding
	CloseAll  key.Binding
	LogLevel  key.Binding
	SubAdd    key.Binding
	SubEdit   key.Binding
	SubDel    key.Binding
	// SubToggle removed
}
```

- [ ] **Step 2: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/model.go
git commit -m "fix: resolve keybinding conflict — SubAdd a→n, remove unused SubToggle"
```

---

### Task 5: Complete handleSubsKeys handlers in TUI update

**Files:**
- Modify: `internal/tui/update.go:258-295`

- [ ] **Step 1: Replace handleSubsKeys with complete implementation**

Replace the entire `handleSubsKeys` function:

```go
func (m Model) handleSubsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	subCount := 0
	if m.cfgMgr != nil {
		subCount = len(m.cfgMgr.Config().Subscriptions)
	}

	switch {
	case key.Matches(msg, Keys.Up):
		if m.subIdx > 0 {
			m.subIdx--
		}

	case key.Matches(msg, Keys.Down):
		if m.subIdx < subCount-1 {
			m.subIdx++
		}

	case key.Matches(msg, Keys.Enter):
		// Set selected subscription as active
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				cur := m.cfgMgr.Config().ActiveSubscription
				if cur == name {
					// Toggle off — use all subscriptions
					m.cfgMgr.SetActiveSubscription("")
					m.subMgr.MergeAndGenerate()
					m.notification = "已切换为使用全部订阅"
				} else {
					m.cfgMgr.SetActiveSubscription(name)
					m.subMgr.MergeAndGenerate()
					m.notification = fmt.Sprintf("已切换激活订阅: %s", name)
				}
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
			}
		}

	case key.Matches(msg, Keys.SubAdd):
		// Prompt for name and URL via notification; actual input requires modal support
		// For now, guide user to CLI: mihomo-cli sub add <name> <url>
		m.notification = "请使用命令行添加: mihomo-cli sub add <名称> <URL>"

	case key.Matches(msg, Keys.SubEdit):
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				m.notification = fmt.Sprintf("请使用命令行编辑: mihomo-cli sub edit %s <新URL>", subs[m.subIdx].Name)
			}
		}

	case key.Matches(msg, Keys.SubDel):
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				m.cfgMgr.RemoveSubscription(name)
				// If we removed the active sub, reset
				if m.cfgMgr.Config().ActiveSubscription == name {
					m.cfgMgr.SetActiveSubscription("")
				}
				m.subMgr.MergeAndGenerate()
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
				m.notification = fmt.Sprintf("已删除订阅 %s", name)
				if m.subIdx >= len(subs)-1 && m.subIdx > 0 {
					m.subIdx--
				}
			}
		}

	case key.Matches(msg, Keys.Update):
		// Update single selected subscription
		if m.subMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				if err := m.subMgr.UpdateSubscription(name); err != nil {
					m.notification = fmt.Sprintf("更新失败: %v", err)
				} else {
					m.notification = fmt.Sprintf("已更新订阅: %s", name)
				}
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
			}
		}
	}
	return m, nil
}
```

- [ ] **Step 2: Update footer help text for subs tab**

In `renderFooter()` (view.go line 149-151), change:

```go
case 4: // 订阅
    keys = append(keys, "↑↓ 导航", "a 添加", "d 删除", "u 更新")
```

to:

```go
case 4: // 订阅
    keys = append(keys, "↑↓ 导航", "Enter 切换激活", "n 添加", "e 编辑", "d 删除", "u 更新")
```

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/tui/update.go internal/tui/view.go
git commit -m "feat: complete TUI subscription tab handlers — switch, add, edit, delete, update"
```

---

### Task 6: Enhance renderSubsTab with active marker

**Files:**
- Modify: `internal/tui/view.go:537-585`

- [ ] **Step 1: Update renderSubsTab to show active subscription**

Replace the `renderSubsTab` function:

```go
// ─── Tab 5: Subscriptions ─────────────────────────────────────

func (m Model) renderSubsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.cfgMgr == nil {
		return PanelStyle.Width(panelW - 2).Render(MutedStyle.Render("配置未加载"))
	}

	subs := m.cfgMgr.Config().Subscriptions
	activeSub := m.cfgMgr.Config().ActiveSubscription

	// Active hint
	activeHint := "全部"
	if activeSub != "" {
		activeHint = activeSub
	}
	b.WriteString(MutedStyle.Render(fmt.Sprintf("  当前激活: %s", BoldStyle.Render(activeHint))))
	b.WriteString("\n\n")

	if len(subs) == 0 {
		empty := MutedStyle.Render("暂无订阅") + "\n\n" +
			HelpKeyStyle.Render("n") + " 添加订阅  " +
			HelpKeyStyle.Render("u") + " 更新全部"
		return PanelStyle.Width(panelW - 2).Render(empty)
	}

	// Header
	b.WriteString(ListHeaderStyle.Render(fmt.Sprintf("  状态  %-16s  %-40s  %s", "名称", "URL", "更新间隔")))
	b.WriteString("\n")
	b.WriteString(DividerStyle.Render("  " + strings.Repeat("─", panelW-8)))
	b.WriteString("\n")

	for i, s := range subs {
		// Active marker
		status := StatusStoppedStyle.Render("○")
		if s.Name == activeSub {
			status = StatusRunningStyle.Render("●")
		}

		urlDisplay := Truncate(s.URL, 38)
		updated := "从未"
		if s.LastUpdated > 0 {
			updated = FormatDuration(time.Now().Unix() - s.LastUpdated) + "前"
		}

		line := fmt.Sprintf("  %s   %-16s  %-40s  %s",
			status, s.Name, urlDisplay, MutedStyle.Render(updated))

		if i == m.subIdx {
			b.WriteString(SelectedStyle.Padding(0, 1).Render(line))
		} else {
			b.WriteString(NormalStyle.Padding(0, 1).Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf("  共 %d 个订阅", len(subs))))

	return PanelStyle.Width(panelW - 2).Render(b.String())
}
```

- [ ] **Step 2: Remove unused "interval" column from header**

The old code had an "更新间隔" column that wasn't populated from the `Subscription` struct's `Interval` field. The new version removes it for simplicity.

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/tui/view.go
git commit -m "feat: show active subscription marker in TUI subs tab"
```

---

### Task 7: Final verification

**Files:** All modified files

- [ ] **Step 1: Full build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: no issues

- [ ] **Step 3: Verify git status**

Run: `git status`
Expected: working tree clean

- [ ] **Step 4: Review full diff**

Run: `git diff main`
Expected: all changes match the design spec
