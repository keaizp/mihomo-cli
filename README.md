# mihomo-cli

Linux 命令行代理管理工具，对标 Clash Verge 的全部功能。基于 mihomo 内核，支持 CLI 子命令和 TUI 交互界面。

## 安装

### 开盒即用：一条命令搞定

mihomo-cli 将 mihomo 内核直接嵌入到二进制中，**无需额外下载或配置**。首次运行时自动解压部署。

```bash
git clone <repo-url> && cd mihomo-cli
make build
sudo cp mihomo-cli /usr/local/bin/
mihomo-cli
```

首次运行输出示例：
```
Extracting embedded mihomo kernel...
Kernel ready: /home/ubuntu/.config/mihomo-cli/mihomo
Starting mihomo...
```

之后直接 `mihomo-cli` 进入 TUI 界面。

**在 Windows/macOS 上交叉编译：**

```bash
# Windows PowerShell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o mihomo-cli ./cmd/mihomo-cli/

# macOS / Linux bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o mihomo-cli ./cmd/mihomo-cli/
```

> **注意**：交叉编译前需要先下载内核文件：`make download-kernel`（或手动放到 `internal/kernel/embedded/mihomo-linux-amd64.gz`）
>
> 在 Windows 上编译默认生成 `.exe`（PE 格式），Linux 无法执行。必须用上面的交叉编译命令。

### 安装到 PATH

```bash
sudo cp mihomo-cli /usr/local/bin/
```

现在可以在任何目录使用：

```bash
mihomo-cli --help
```

### 验证安装

```bash
which mihomo-cli        # /usr/local/bin/mihomo-cli
mihomo-cli --help       # 打印命令列表
```

### 内核管理（高级）

正常情况下不需要关心内核，mihomo-cli 自动管理。以下命令用于特殊场景（使用自定义内核、更换版本等）：

```bash
mihomo-cli kernel path                           # 查看内核安装路径
mihomo-cli kernel install --local ./mihomo       # 使用本地内核替换
mihomo-cli kernel install --url <镜像地址>        # 从镜像下载替换
```

### 依赖

- Go 1.22+（仅编译时需要）
- Linux amd64（嵌入内核仅支持此平台）
- GNOME 桌面环境（系统代理功能需要 gsettings）

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

### 内核管理 `kernel`

正常情况下内核已嵌入二进制，开盒即用。以下命令用于手动替换内核：

```bash
# 查看内核安装路径
mihomo-cli kernel path

# 使用本地内核替换
mihomo-cli kernel install --local ~/Downloads/mihomo-linux-amd64-v1.18.10.gz

# 从镜像下载替换
mihomo-cli kernel install --url https://mirror.example.com/mihomo.gz
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
mihomo-cli                              # 首次运行自动部署内核并进入 TUI
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

- mihomo 内核已嵌入二进制，首次运行自动解压到 `~/.config/mihomo-cli/mihomo`，无需手动安装
- 非 Linux/amd64 平台不支持嵌入内核，需手动安装：`mihomo-cli kernel install --local <path>`
- 系统代理设置仅支持 GNOME 桌面环境（通过 gsettings）
- mihomo 内核以子进程方式运行，退出 mihomo-cli 后内核继续运行。如需停止请使用 `mihomo-cli service stop`

## 常见问题

### `command not found`

二进制不在 PATH 环境变量中。解决方法：

```bash
# 方案 A：安装到 PATH 目录（推荐）
sudo cp mihomo-cli /usr/local/bin/
mihomo-cli --help

# 方案 B：在当前目录带路径运行
./mihomo-cli --help
```

### `Permission denied`

Linux 文件系统不会自动给新文件执行权限。解决方法：

```bash
chmod +x ./mihomo-cli
./mihomo-cli --help
```

### 编译后 `go build` 生成的二进制也无法执行

同上，某些场景下 `go build` 可能不保留执行位：

```bash
go build -o mihomo-cli ./cmd/mihomo-cli/
chmod +x ./mihomo-cli    # 确保有执行权限
```

### 内核相关问题

**正常运行不需要关心内核** — 它已嵌入在二进制中，首次运行自动部署。

如果需要手动替换内核（更换版本等）：

```bash
# 查看当前内核路径
mihomo-cli kernel path

# 使用本地文件替换
mihomo-cli kernel install --local ./mihomo

# 或从镜像下载
mihomo-cli kernel install --url https://mirror.example.com/mihomo-linux-amd64.gz
```

### `MZ... not found` / `Syntax error: Unterminated quoted string`

这是因为运行了 **Windows 版本的二进制**。在 Windows 上编译会生成 `.exe`（PE 格式），Linux 无法执行。解决方案：

```bash
# 在 Windows 上交叉编译（PowerShell）
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mihomo-cli ./cmd/mihomo-cli/

# 或者在 Linux 上直接重新编译
go build -o mihomo-cli ./cmd/mihomo-cli/
```

生成的 Linux 二进制没有 `.exe` 后缀，传过去 `chmod +x` 后即可运行。
