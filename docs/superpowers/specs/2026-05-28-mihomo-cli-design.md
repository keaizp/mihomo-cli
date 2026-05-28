# Mihomo CLI — Design Spec

## Overview

A Linux CLI tool (`mihomo-cli`) that replaces Clash Verge with first-class command-line and TUI experiences. Single Go binary, auto-downloads mihomo kernel, manages subscriptions, proxies, modes, and connections.

## Tech Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Language | Go 1.22+ | mihomo is Go, reuse ecosystem, single static binary |
| CLI Framework | cobra | Standard, subcommand support, completion |
| TUI Framework | bubbletea + lipgloss | Elm-like architecture, vibrant ecosystem |
| Kernel | mihomo | Most active Clash fork, Clash Verge default |
| Kernel Dist | Auto-download from GitHub releases | UX parity with Clash Verge |

## Architecture

```
CLI Subcommands (cobra)  │  TUI (bubbletea)
         └────────────────┬────────────────┘
                   Core Logic Layer
   ┌──────────┬──────────┬──────────┬──────────┐
   │ Subscription │ Config  │ Mihomo  │ Kernel  │
   │ Manager    │ Manager │ API     │ Manager │
   └──────────┴──────────┴──────────┴──────────┘
                          │
                   Mihomo Kernel (child process)
                   REST API :9090 | SOCKS :1080
```

Three layers:
- **Interaction**: cobra subcommands + bubbletea TUI, sharing core logic
- **Core**: subscription parsing, config I/O, mihomo API client, kernel lifecycle
- **Runtime**: mihomo runs as a child process, tool controls it via REST API

## CLI Command Tree

```
mihomo-cli                          # No args → launch TUI

# Subscription management
mihomo-cli sub add <name> <url>
mihomo-cli sub remove <name>
mihomo-cli sub update [name]
mihomo-cli sub list

# Proxy mode
mihomo-cli mode set <rule|global|direct|script>
mihomo-cli mode show

# Node management
mihomo-cli proxy list [--group <name>]
mihomo-cli proxy set <group> <node>
mihomo-cli proxy test [--group <name>]
mihomo-cli proxy info <node>

# Service control
mihomo-cli service start|stop|restart|status|logs

# Connection management
mihomo-cli conn list
mihomo-cli conn close <id>

# Config
mihomo-cli config edit|show|reload
```

## TUI Layout

Three-panel layout:

```
┌─ Proxies (left) ──────┐ ┌─ Node Info (top-right) ─┐
│ [GLOBAL]          ◄►   │ │ Name/Type/Addr/Delay     │
│ ├─ HK-01  32ms         │ │ Traffic stats            │
│ ├─ JP-02 120ms         │ └──────────────────────────┘
│ [Streaming]        ◄►  │ ┌─ Log (bottom-right) ────┐
│ ├─ Netflix 45ms        │ │ Real-time event stream   │
│ └─ DIRECT              │ └──────────────────────────┘
└────────────────────────┘
```

Keybindings: Tab (switch group), arrows (navigate), Enter (select), t (latency test), r (reload), s (start/stop service), u (update subscriptions), q (quit).

## Core Module Responsibilities

### Subscription Manager
- Parse Clash-compatible subscription URLs (base64 → YAML)
- Fetch and merge subscriptions into mihomo config
- Auto-update on schedule (configurable interval)
- Deduplicate nodes across subscriptions

### Config Manager
- Manage mihomo config file at `$XDG_CONFIG_HOME/mihomo-cli/config.yaml`
- Merge subscription nodes with user overrides
- Write config to mihomo's working directory
- Handle config reload without restart

### Mihomo API Client
- HTTP client targeting mihomo REST API (`http://127.0.0.1:9090`)
- Operations: get proxies, switch proxy, test latency, get connections, close connection
- Handle API unavailability gracefully

### Kernel Manager
- Download mihomo binary from GitHub on first run
- Start/stop/restart mihomo as child process
- Health check via API endpoint
- Log capture and rotation
- Version management and self-update

### System Proxy
- Set system HTTP/HTTPS/SOCKS proxy via gsettings / environment variables
- TUN mode support (via mihomo)
- Restore on exit

## Data Flow

```
Subscribe URL ──► Fetch ──► Decode base64 ──► Parse YAML
    └──► Merge nodes ──► Write config.yaml
    └──► Reload mihomo ──► Nodes available

User switches proxy:
  CLI/TUI ──► API Client ──► PUT /proxies/{group}
  └──► mihomo handles routing
  └──► System proxy updated (if applicable)
```

## Configuration

Stored at `$XDG_CONFIG_HOME/mihomo-cli/`:
- `config.yaml` — user preferences, subscriptions, overrides
- `profiles/` — per-subscription cached configs
- `mihomo/` — mihomo working directory (config, database)

## Error Handling

- API unavailable: retry with backoff, surface to TUI as status banner
- Subscription fetch failure: keep last-known-good config, notify user
- Kernel download failure: prompt user to retry or specify binary path
- Invalid config: validate before applying, reject with line-specific error

## Testing Strategy

- Unit tests for subscription parsing, config merge logic
- Integration tests with a mihomo instance in test mode
- TUI tested via bubbletea's built-in test framework

## Non-Goals

- Windows/macOS support (Linux-only initially)
- Built-in rule editor (use `config edit` + external editor)
- Graphical dashboard
- Plugin system
