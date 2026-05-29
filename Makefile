.PHONY: build clean update-kernel install uninstall install-completions

# Install paths (override with: make install PREFIX=/usr)
PREFIX  ?= /usr/local
BINDIR  ?= $(PREFIX)/bin
DATADIR ?= /var/lib/mihomo-cli

build:
	go build -ldflags="-s -w" -o mihomo-cli ./cmd/mihomo-cli/
	@echo "✓ 构建完成: ./mihomo-cli"

# update-kernel downloads a newer mihomo release (optional, for developers).
# The repo already ships with a working kernel embedded — no download needed for build.
MIHOMO_VERSION := v1.18.10
EMBED_DIR     := internal/kernel/embedded
EMBED_FILE    := $(EMBED_DIR)/mihomo-linux-amd64.gz
MIHOMO_URL    := https://github.com/MetaCubeX/mihomo/releases/download/$(MIHOMO_VERSION)/mihomo-linux-amd64-$(MIHOMO_VERSION).gz

update-kernel:
	@mkdir -p $(EMBED_DIR)
	curl -fL -o $(EMBED_FILE) $(MIHOMO_URL)
	@echo "✓ 内核已更新: $(EMBED_FILE)"

# ─── Install / Uninstall ─────────────────────────────────────

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 mihomo-cli $(DESTDIR)$(BINDIR)/mihomo-cli
	install -d $(DESTDIR)$(DATADIR)
	@echo "✓ mihomo-cli 安装成功"
	@echo "  二进制: $(DESTDIR)$(BINDIR)/mihomo-cli"
	@echo "  数据目录: $(DATADIR)"

uninstall:
	-$(DESTDIR)$(BINDIR)/mihomo-cli service stop 2>/dev/null || true
	rm -f $(DESTDIR)$(BINDIR)/mihomo-cli
	rm -rf $(DESTDIR)$(DATADIR)
	@echo "✓ mihomo-cli 已卸载"
	@echo "  用户配置 ~/.config/mihomo-cli/ 已保留，如需删除: rm -rf ~/.config/mihomo-cli"

install-completions:
	$(DESTDIR)$(BINDIR)/mihomo-cli completion bash | sudo tee /etc/bash_completion.d/mihomo-cli > /dev/null 2>&1 || true
	$(DESTDIR)$(BINDIR)/mihomo-cli completion zsh  | sudo tee /usr/local/share/zsh/site-functions/_mihomo-cli > /dev/null 2>&1 || true
	$(DESTDIR)$(BINDIR)/mihomo-cli completion fish | sudo tee /usr/local/share/fish/completions/mihomo-cli.fish > /dev/null 2>&1 || true
	@echo "✓ Shell 补全已安装，重新打开终端生效"

clean:
	rm -f mihomo-cli mihomo-cli.exe
