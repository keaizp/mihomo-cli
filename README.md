# mihomo-cli

Linux 命令行代理管理工具，对标 Clash Verge 的全部功能。基于 mihomo 内核，支持 CLI 子命令和 TUI 交互界面。

## 安装

### 1. 安装 mihomo-cli 本体

**在 Linux 上直接编译：**

```bash
git clone <repo-url> && cd mihomo-cli
go build -o mihomo-cli ./cmd/mihomo-cli/
chmod +x ./mihomo-cli
sudo cp mihomo-cli /usr/local/bin/
```

**在 Windows/macOS 上交叉编译 Linux 版本：**

```bash
# Windows PowerShell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mihomo-cli ./cmd/mihomo-cli/

# macOS / Linux bash
GOOS=linux GOARCH=amd64 go build -o mihomo-cli ./cmd/mihomo-cli/
```

> **注意**：在 Windows 上编译默认生成 `.exe` 文件，那是 Windows PE 格式，Linux 无法执行。必须用上面的交叉编译命令。

### 2. 安装 mihomo 内核（必需）

mihomo-cli 本身不包含代理内核，需要单独准备。**程序不会自动联网下载**（因为需要用代理的人网络本来就有问题）。

**方案 A：手动下载（推荐，最可靠）**

在你网络最好的时候，用浏览器或其他方式下载 mihomo 二进制，然后安装：

```bash
# 1. 查看期望的安装路径
mihomo-cli kernel path
# 输出: /home/ubuntu/.config/mihomo-cli/mihomo

# 2. 把下载好的 mihomo 复制过去
mihomo-cli kernel install --local ./下载好的mihomo
```

mihomo 官方下载地址：[github.com/MetaCubeX/mihomo/releases](https://github.com/MetaCubeX/mihomo/releases)  
需要下载 `mihomo-linux-amd64-<version>.gz`（文件名以 `.gz` 结尾的），无需手动解压。

**方案 B：通过镜像下载**

如果你有可用的镜像站：

```bash
mihomo-cli kernel install --url https://your-mirror.example.com/mihomo-linux-amd64.gz
```

**方案 C：直接下载（需要 GitHub 可达）**

```bash
mihomo-cli kernel install
```

会显示下载进度条，从 GitHub 直接拉取。
./mihomo-cli --help
```

### 2. 安装到 PATH（推荐）

要像系统命令一样在任何目录直接使用，需要把二进制复制到 PATH 目录：

```bash
# 安装到 /usr/local/bin（对所有用户生效）
sudo cp mihomo-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/mihomo-cli

# 或者安装到当前用户的 ~/.local/bin
mkdir -p ~/.local/bin
cp mihomo-cli ~/.local/bin/
# 确保 ~/.local/bin 在 PATH 中：export PATH="$HOME/.local/bin:$PATH"
```

安装后可以直接使用：

```bash
mihomo-cli --help
```

> **为什么出现 `command not found`？**  
> Linux 只会搜索 PATH 环境变量中的目录（`echo $PATH` 查看）。  
> 如果二进制不在这些目录里，就需要用 `./mihomo-cli` 带路径运行，或者安装到 PATH 目录。

### 3. 验证安装

```bash
which mihomo-cli   # 应该输出 /usr/local/bin/mihomo-cli
mihomo-cli --help   # 应该打印命令列表
```

### 依赖

- Go 1.22+（仅编译时需要）
- Linux（系统代理功能依赖 GNOME gsettings）
- mihomo 内核需手动安装，参见上方第 2 步

## 快速开始

```bash
# 0. 安装 mihomo 内核（首次必须）
mihomo-cli kernel install --local ./mihomo

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

```bash
# 查看内核期待安装路径
mihomo-cli kernel path

# 从 GitHub 下载内核（需要网络可达 GitHub）
mihomo-cli kernel install

# 从镜像下载
mihomo-cli kernel install --url https://mirror.example.com/mihomo.gz

# 从本地文件安装（最常用：手动下载后安装）
mihomo-cli kernel install --local ~/Downloads/mihomo-linux-amd64-v1.18.10.gz
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

- mihomo 内核需手动安装，不会自动下载。安装路径用 `mihomo-cli kernel path` 查看
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

### 下载内核失败

检查网络是否可达 GitHub：

```bash
curl -I https://github.com/MetaCubeX/mihomo/releases
```

如果网络受限，可以手动下载 mihomo 二进制放到 `~/.config/mihomo-cli/mihomo` 并确保有执行权限。

### `MZ... not found` / `Syntax error: Unterminated quoted string`

这是因为运行了 **Windows 版本的二进制**。在 Windows 上编译会生成 `.exe`（PE 格式），Linux 无法执行。解决方案：

```bash
# 在 Windows 上交叉编译（PowerShell）
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mihomo-cli ./cmd/mihomo-cli/

# 或者在 Linux 上直接重新编译
go build -o mihomo-cli ./cmd/mihomo-cli/
```

生成的 Linux 二进制没有 `.exe` 后缀，传过去 `chmod +x` 后即可运行。
