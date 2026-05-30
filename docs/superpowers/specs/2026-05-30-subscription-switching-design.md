# Subscription Switching Design

## Summary

Add active-subscription switching (single-selection mode), fix TUI keybinding conflicts, and complete the subscription tab interaction handlers.

## Data Model

### AppConfig (`internal/cfg/manager.go`)

New field:

```yaml
active_subscription: ""  # name of active sub; empty = merge all (backwards-compat)
```

New methods on `*Manager`:
- `SetActiveSubscription(name string) error` — validate name exists, set, save
- `UpdateSubscription(name, url string) error` — find by name, update URL, save

### Subscription struct — unchanged

No `Enabled` field needed for single-selection mode.

## Merge Logic (`internal/subscription/manager.go`)

`MergeAndGenerate`:
- If `ActiveSubscription` is non-empty, only merge proxies from that subscription
- If empty, merge all (backwards-compatible)

## CLI (`internal/cli/sub.go`)

New commands:

| Command | Description |
|---------|-------------|
| `sub switch <name>` | Set active subscription |
| `sub edit <name> <url>` | Update subscription URL |

`sub update` changed: when given a name, after fetch + save, if that sub is the active one, regenerate mihomo config.

## TUI

### Keybinding fixes (`internal/tui/model.go`)

| Binding | Old Key | New Key | Reason |
|---------|---------|---------|--------|
| SubAdd | `a` | `n` | Conflict with TestAll |
| SubToggle | `space` | **removed** | Replaced by Enter for switching |

### SubEdit binding: change key from `e` to `e` (keep), add handler.

### handleSubsKeys (`internal/tui/update.go`)

Complete handlers:

| Key | Action |
|-----|--------|
| `n` | Add subscription (prompt for name + URL) |
| `e` | Edit selected subscription URL |
| `d` | Delete selected subscription (existing) |
| `Enter` | Set selected subscription as active |
| `u` | Update selected subscription (single) |
| `↑↓` | Navigate (existing) |

### renderSubsTab (`internal/tui/view.go`)

- Active subscription: green `●` marker
- Inactive: `○` marker
- Header shows "当前激活: <name>" or "当前激活: 全部"

## Files Changed

| File | Changes |
|------|---------|
| `internal/cfg/manager.go` | +ActiveSubscription, +SetActiveSubscription, +UpdateSubscription |
| `internal/subscription/manager.go` | MergeAndGenerate: filter by active sub |
| `internal/cli/sub.go` | +switch cmd, +edit cmd, update cmd: regenerate config for active sub |
| `internal/tui/model.go` | SubAdd key `a`→`n`, remove SubToggle |
| `internal/tui/update.go` | Complete handleSubsKeys with all handlers |
| `internal/tui/view.go` | renderSubsTab: show active marker, active hint |
