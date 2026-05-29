# mihomo-cli TUI Redesign

## Goal

Redesign the TUI interface to match ccswitch-level polish with full Chinese localization, tab-based multi-view architecture, and comprehensive proxy management features.

## Architecture: Tab-Based Multi-View

The TUI uses a tab bar with 5 views, switched via number keys `1-5`:

| Key | Tab | Purpose |
|-----|-----|---------|
| `1` | 代理 Proxies | Main proxy list, latency, traffic, search |
| `2` | 连接 Connections | Active connections, close single/all |
| `3` | 日志 Logs | Real-time kernel logs with level filter |
| `4` | 规则 Rules | Routing rules display |
| `5` | 订阅 Subs | Subscription management (add/del/toggle/update) |

A shared header bar shows app name, version, run status, mode, and real-time traffic. A shared footer shows context-sensitive keybindings in Chinese.

## Views

### Tab 1: 代理 Proxies
- List all proxy groups with collapsible headers (Space to toggle)
- Show nodes with: selection marker, name, color-coded latency, traffic bar, upload/download
- Latency color: green (<200ms), yellow (<500ms), red (>=500ms), gray (untested)
- `t` test current node, `T` test ALL nodes in current group
- `/` enter search filter mode, ESC to clear
- `Enter` to switch proxy
- Group header shows node count and current selection

### Tab 2: 连接 Connections
- Table: protocol, host, proxy chain, upload, download
- `d` close selected connection, `X` close all connections
- Auto-refresh every 2 seconds
- Show total connection count in footer

### Tab 3: 日志 Logs
- Streaming log display from mihomo kernel
- `l` cycle log level filter: all → info → warning → error → debug
- Auto-scroll toggle with `s`
- Color-coded log levels

### Tab 4: 规则 Rules
- Read-only display of routing rules from API
- Show: type, payload, proxy target
- Scrollable list

### Tab 5: 订阅 Subs
- List subscriptions with status (enabled/disabled), URL, last update time
- `a` add subscription (prompt for name + URL)
- `e` edit selected subscription
- `d` delete selected subscription
- `Space` toggle enable/disable
- `u` update all subscriptions

## Global Keys (all views)

| Key | Action |
|-----|--------|
| `1-5` | Switch tab |
| `↑↓/jk` | Navigate list |
| `q/ESC/Ctrl+C` | Quit |
| `m` | Cycle mode: rule→global→direct→script |
| `r` | Reload mihomo config |
| `u` | Update all subscriptions |

## Header Bar

```
🚀 mihomo-cli vX.Y.Z │ 🟢 运行中 │ 规则模式 │ ↑ 2.1MB/s ↓ 12.8MB/s
```

- Run status: green dot + 运行中, red dot + 已停止, yellow dot + 启动中
- Mode: 规则模式 / 全局模式 / 直连模式 / 脚本模式
- Traffic: real-time upload/download rates, refreshed every 1s

## Data Flow

- `tickMsg` every 1s: fetch traffic stats, update header
- `tickMsg` every 2s: fetch proxies (tab 1), fetch connections (tab 2), fetch logs (tab 3)
- Delay test results stored in `map[string]int` on Model, displayed inline
- Subscription state read from `cfg.Manager`, changes persisted immediately

## Model State Additions

```
tabIdx        int               // 0-4 current tab
delayResults  map[string]int    // node → latency in ms
connections   []api.Connection  // active connections
logs          []string          // recent log lines
logLevel      string            // log filter level
searchQuery   string            // search/filter text
collapsed     map[string]bool   // collapsed groups
traffic       struct{Up, Down int64}  // real-time rates
cfgMgr        *cfg.Manager      // for mode and subs management
```

## Files Changed

- `internal/tui/model.go` — expanded Model, KeyMap with Chinese help
- `internal/tui/view.go` — tab routing, 5 view renderers, header, footer
- `internal/tui/update.go` — message handling for all views
- `internal/tui/styles.go` — comprehensive style system
- `internal/cli/root.go` — pass cfgMgr to TUI model

## Non-Goals

- Mouse support (focus on keyboard navigation)
- Custom themes (hardcode a good default)
- Multi-language (Chinese only for UI text)
