# mihomo-cli

Linux 代理管理工具，CLI + TUI 双模式，对标 Clash Verge 全部功能。

## 安装

```bash
git clone <repo-url> && cd mihomo-cli
make build
sudo make install
```

安装后在任何目录直接使用：

```bash
mihomo-cli                 # 进入 TUI 界面
mihomo-cli --help          # 查看所有命令
```

**可选：**

```bash
sudo make install-completions          # Shell 自动补全（bash/zsh/fish）
sudo cp mihomo-cli.service /etc/systemd/system/  # systemd 服务（开机自启）
sudo systemctl enable --now mihomo-cli
```

**卸载：**

```bash
sudo make uninstall
```

## 快速上手

```bash
# 1. 添加订阅
mihomo-cli sub add 我的订阅 "https://your-sub-url"

# 2. 更新订阅
mihomo-cli sub update

# 3. 启动代理
mihomo-cli service start

# 4. 进入 TUI 选节点（或用 proxy set 命令行切换）
mihomo-cli
```

## 命令速查

### 服务控制

| 命令 | 说明 |
|------|------|
| `mihomo-cli service start` | 启动代理 |
| `mihomo-cli service stop` | 停止代理 |
| `mihomo-cli service restart` | 重启代理 |
| `mihomo-cli service status` | 查看状态 |
| `mihomo-cli service logs` | 查看日志 |

### 节点与模式

| 命令 | 说明 |
|------|------|
| `mihomo-cli proxy list` | 查看所有节点 |
| `mihomo-cli proxy set <组> <节点>` | 切换节点 |
| `mihomo-cli proxy test` | 测试延迟 |
| `mihomo-cli mode set rule` | 规则模式（推荐） |
| `mihomo-cli mode set global` | 全局模式 |
| `mihomo-cli mode set direct` | 直连模式 |

### 订阅管理

| 命令 | 说明 |
|------|------|
| `mihomo-cli sub add <名称> <URL>` | 添加订阅 |
| `mihomo-cli sub list` | 列出订阅 |
| `mihomo-cli sub update` | 更新全部订阅 |
| `mihomo-cli sub remove <名称>` | 删除订阅 |

### 连接与配置

| 命令 | 说明 |
|------|------|
| `mihomo-cli conn list` | 查看活跃连接 |
| `mihomo-cli conn close <ID>` | 关闭连接 |
| `mihomo-cli config show` | 查看配置 |
| `mihomo-cli config edit` | 编辑配置 |
| `mihomo-cli config reload` | 重载配置 |

## TUI 交互界面

直接运行 `mihomo-cli`（无参数）进入全屏交互界面。

**5 个视图，数字键 1-5 切换：**

| 按键 | 视图 | 功能 |
|------|------|------|
| `1` | 代理 | 浏览节点、切换代理、测速、搜索 |
| `2` | 连接 | 查看活跃连接、关闭连接 |
| `3` | 日志 | 实时日志、按级别过滤 |
| `4` | 规则 | 路由规则列表 |
| `5` | 订阅 | 管理订阅 |

**常用快捷键：**

| 按键 | 功能 |
|------|------|
| `↑↓` / `j` `k` | 导航 |
| `Enter` | 切换代理 |
| `Tab` | 切换代理组 |
| `t` | 测速当前节点 |
| `T` | 全组测速 |
| `/` | 搜索节点 |
| `空格` | 折叠代理组 |
| `m` | 切换模式 |
| `r` | 重载配置 |
| `u` | 更新订阅 |
| `q` | 退出 |

## 使用代理

mihomo 启动后监听 **混合端口 7890**（HTTP + SOCKS5）。

**终端：**

```bash
export http_proxy=http://127.0.0.1:7890 https_proxy=http://127.0.0.1:7890
unset http_proxy https_proxy    # 取消
```

可加入 `~/.bashrc` 永久生效，或绑定别名方便开关。

**浏览器：** 安装 SwitchyOmega 插件，添加 HTTP 代理 `127.0.0.1:7890`。

**其他程序：** 在各自的网络设置中配置 HTTP 代理 `127.0.0.1:7890`。

## 验证代理

```bash
# 对比代理前后的 IP
curl -s https://ip.sb                    # 直连 IP
curl -x http://127.0.0.1:7890 -s https://ip.sb   # 代理 IP

# 测试节点延迟
mihomo-cli proxy test
```

## 配置文件

`~/.config/mihomo-cli/config.yaml`：

```yaml
core:
  http_port: 7890
  socks_port: 7891
  api_port: 9090
  log_level: info

mode: rule

subscriptions:
  - name: 我的订阅
    url: https://example.com/subscribe
    interval: 86400
```

### 目录结构

```
~/.config/mihomo-cli/          # 用户配置（订阅、模式）
/var/lib/mihomo-cli/           # 守护进程数据（内核、日志、PID）
```

## 常见问题

**`command not found`** — 没有安装到 PATH。运行 `sudo make install` 或手动 `sudo cp mihomo-cli /usr/local/bin/`。

**日志文件不存在** — 检查 mihomo 是否启动：`mihomo-cli service status`。如果没启动，先 `mihomo-cli service start`。

**切了节点但 IP 没变** — 确认应用正在使用代理端口 7890，或切换到全局模式：`mihomo-cli mode set global`。

**非 Linux/amd64 平台** — 内核仅嵌入在 Linux amd64 版本中。其他平台需手动安装内核：`mihomo-cli kernel install --local <路径>`。
