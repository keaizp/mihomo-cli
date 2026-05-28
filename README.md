# mihomo-cli

Linux 命令行代理管理工具，对标 Clash Verge 的全部功能。基于 mihomo 内核，支持 CLI 子命令和 TUI 交互界面。

## 安装

### 从源码编译

```bash
git clone <repo-url> && cd mihomo-cli
go build -o mihomo-cli ./cmd/mihomo-cli/
sudo mv mihomo-cli /usr/local/bin/
```

### 依赖

- Go 1.22+
- Linux（系统代理功能依赖 GNOME gsettings）

首次运行会自动从 GitHub 下载 mihomo 内核，无需手动安装。

## 快速开始

```bash
# 1. 添加订阅
mihomo-cli sub add my-sub "https://your-subscription-url"

# 2. 更新订阅（拉取节点列表）
mihomo-cli sub update my-sub

# 3. 查看节点
mihomo-cli proxy list

# 4. 切换节点
mihomo-cli proxy set GLOBAL "🇭🇰 HK-01"

# 5. 测试延迟
mihomo-cli proxy test

# 6. 或者直接进入 TUI 交互界面
mihomo-cli
```

## 命令参考

### 订阅管理 `sub`

```bash
# 添加订阅
mihomo-cli sub add <名称> <URL>
mihomo-cli sub add my-vpn "https://example.com/subscribe?token=xxx"

# 列出所有订阅
mihomo-cli sub list

# 更新指定订阅
mihomo-cli sub update my-vpn

# 更新全部订阅
mihomo-cli sub update

# 删除订阅
mihomo-cli sub remove my-vpn
```

### 代理模式 `mode`

```bash
# 查看当前模式
mihomo-cli mode show

# 切换模式
mihomo-cli mode set rule     # 规则模式（推荐，按规则分流）
mihomo-cli mode set global   # 全局模式（所有流量走代理）
mihomo-cli mode set direct   # 直连模式（所有流量不走代理）
mihomo-cli mode set script   # 脚本模式
```

### 节点管理 `proxy`

```bash
# 列出所有代理组和节点（● 标记当前选中）
mihomo-cli proxy list

# 按分组列出
mihomo-cli proxy list --group GLOBAL

# 切换节点
mihomo-cli proxy set <分组> <节点名>
mihomo-cli proxy set GLOBAL "🇭🇰 HK-01"

# 测试所有节点延迟
mihomo-cli proxy test

# 测试指定节点延迟
mihomo-cli proxy test "🇭🇰 HK-01"

# 查看节点详情
mihomo-cli proxy info "🇭🇰 HK-01"
```

### 服务控制 `service`

```bash
# 启动 mihomo 内核
mihomo-cli service start

# 停止内核
mihomo-cli service stop

# 重启内核
mihomo-cli service restart

# 查看运行状态
mihomo-cli service status
# 输出: mihomo: running / stopped / starting

# 查看日志
mihomo-cli service logs
```

### 连接管理 `conn`

```bash
# 查看当前所有连接
mihomo-cli conn list
# 输出: ID | HOST | NETWORK | RULE | UPLOAD | DOWNLOAD

# 关闭指定连接
mihomo-cli conn close <连接ID>
```

### 配置管理 `config`

```bash
# 查看当前配置
mihomo-cli config show
# 输出: Mode, HTTP/SOCKS/API 端口, 订阅数量

# 用编辑器打开配置文件
mihomo-cli config edit
# 使用 $EDITOR 环境变量指定的编辑器，默认 vim

# 重载配置（无需重启内核）
mihomo-cli config reload
```

## TUI 交互界面

直接运行 `mihomo-cli`（无参数）进入全屏 TUI：

```
┌── Proxies ────────────────────┐  ┌── Details ───────────┐
│                               │  │ Name: HK-01           │
│  [GLOBAL]                     │  │ Type: vmess           │
│   ● 🇭🇰 HK-01    32ms         │  │ Delay: 32ms           │
│     🇯🇵 JP-02   120ms         │  │                       │
│     🇸🇬 SG-03   250ms         │  └───────────────────────┘
│                               │
│  [Streaming]                  │  ↑↓ navigate  tab group
│   ● 🇭🇰 Netflix  45ms         │  enter switch
│     🇺🇸 Disney   180ms         │  t test  r reload
│                               │  u update  q quit
│  [Direct]                     │
│   ● DIRECT                    │
│                               │
├───────────────────────────────┤
│  mihomo: running     mode: N/A│
└───────────────────────────────┘
```

### 键盘快捷键

| 按键 | 功能 |
|---:|---|
| `↑` `↓` / `j` `k` | 上下移动光标 |
| `Tab` | 切换到下一个代理组 |
| `Enter` | 切换到选中的节点 |
| `t` | 测试当前节点延迟 |
| `r` | 重载配置文件 |
| `u` | 更新所有订阅 |
| `q` / `Ctrl+C` | 退出 |

## 配置文件

配置文件位于 `$XDG_CONFIG_HOME/mihomo-cli/config.yaml`（默认 `~/.config/mihomo-cli/config.yaml`）：

```yaml
core:
  http_port: 7890       # HTTP 代理端口
  socks_port: 7891      # SOCKS5 代理端口
  mixed_port: 7892      # 混合端口（HTTP+SOCKS）
  api_port: 9090        # REST API 端口
  allow_lan: false      # 是否允许局域网连接
  log_level: info       # 日志级别：debug/info/warning/error

mode: rule              # 默认模式：rule/global/direct/script

subscriptions:
  - name: my-sub
    url: https://example.com/subscribe
    interval: 86400    # 自动更新间隔（秒），0 表示手动更新
    last_updated: 0

user_proxies: []        # 自定义节点（YAML 格式）
user_rules: []           # 自定义规则
```

### 目录结构

```
~/.config/mihomo-cli/
├── config.yaml          # 应用配置
├── profiles/            # 订阅缓存
│   └── my-sub.yaml      # 每个订阅的节点数据
└── mihomo/              # mihomo 内核工作目录
    └── config.yaml      # 自动生成的 mihomo 配置
```

## 环境变量

| 变量 | 说明 | 默认值 |
|---|---|---|
| `XDG_CONFIG_HOME` | 配置目录 | `~/.config` |
| `EDITOR` | config edit 使用的编辑器 | `vim` |

## 系统代理

在 GNOME 桌面环境下，可以通过 `internal/sysproxy` 模块设置系统代理。代理开关后会自动调用 `gsettings` 设置 HTTP/HTTPS 代理。

## 常见操作流程

### 初次使用

```bash
mihomo-cli sub add main "https://your-sub-url"
mihomo-cli sub update main
mihomo-cli proxy test
mihomo-cli proxy set GLOBAL <延迟最低的节点>
```

### 日常使用

```bash
# 启动代理
mihomo-cli service start

# 进入 TUI 浏览和切换节点
mihomo-cli

# 更新订阅获取最新节点
mihomo-cli sub update

# 停止代理
mihomo-cli service stop
```

### 脚本化 / 自动化

```bash
# 开机自启（添加到 systemd 或 ~/.profile）
mihomo-cli service start

# 定时更新订阅（添加到 crontab）
# 每天 6:17 更新全部订阅
17 6 * * * /usr/local/bin/mihomo-cli sub update

# 快捷切换模式（绑定别名）
alias proxy-on='mihomo-cli mode set rule'
alias proxy-off='mihomo-cli mode set direct'
alias proxy-global='mihomo-cli mode set global'
```

## 注意事项

- 首次运行会自动下载 mihomo 内核（~15MB），需要网络连接
- 内核二进制存放在 `$XDG_CONFIG_HOME/mihomo-cli/mihomo`
- 系统代理设置仅支持 GNOME 桌面环境（通过 gsettings）
- mihomo 内核以子进程方式运行，退出 mihomo-cli 后内核继续运行。如需停止请使用 `mihomo-cli service stop`
