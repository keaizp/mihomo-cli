.PHONY: build clean download-kernel install uninstall install-completions

MIHOMO_VERSION := v1.18.10
EMBED_DIR     := internal/kernel/embedded
EMBED_FILE    := $(EMBED_DIR)/mihomo-linux-amd64.gz
MIHOMO_URL    := https://github.com/MetaCubeX/mihomo/releases/download/$(MIHOMO_VERSION)/mihomo-linux-amd64-$(MIHOMO_VERSION).gz

# Install paths (override with: make install PREFIX=/usr)
PREFIX  ?= /usr/local
BINDIR  ?= $(PREFIX)/bin
DATADIR ?= /var/lib/mihomo-cli

build: download-kernel
	go build -ldflags="-s -w" -o mihomo-cli ./cmd/mihomo-cli/
	@echo "Build complete: ./mihomo-cli"

download-kernel: $(EMBED_FILE)

$(EMBED_FILE):
	@mkdir -p $(EMBED_DIR)
	@echo "Downloading mihomo $(MIHOMO_VERSION)..."
	@curl -fL -o $(EMBED_FILE) $(MIHOMO_URL) || ( \
		echo ""; \
		echo "Failed to download mihomo from GitHub."; \
		echo "Download it manually and place it at $(EMBED_FILE):"; \
		echo "  $(MIHOMO_URL)"; \
		exit 1 \
	)
	@echo "Kernel downloaded to $(EMBED_FILE)"

# ─── Install / Uninstall ─────────────────────────────────────

install: build
	@echo "Installing mihomo-cli..."
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 mihomo-cli $(DESTDIR)$(BINDIR)/mihomo-cli
	install -d $(DESTDIR)$(DATADIR)
	@echo ""
	@echo "mihomo-cli installed to $(DESTDIR)$(BINDIR)/mihomo-cli"
	@echo "Data directory: $(DATADIR)"
	@echo ""
	@echo "Next steps:"
	@echo "  mihomo-cli              # Start TUI (auto-deploys kernel on first run)"
	@echo "  mihomo-cli --help       # Show all commands"
	@echo ""
	@echo "Install shell completions:"
	@echo "  make install-completions"

uninstall:
	@echo "Uninstalling mihomo-cli..."
	-$(DESTDIR)$(BINDIR)/mihomo-cli service stop 2>/dev/null || true
	rm -f $(DESTDIR)$(BINDIR)/mihomo-cli
	rm -rf $(DESTDIR)$(DATADIR)
	@echo ""
	@echo "mihomo-cli removed from $(DESTDIR)$(BINDIR)/"
	@echo "Daemon data removed from $(DATADIR)/"
	@echo ""
	@echo "User config at ~/.config/mihomo-cli/ was kept."
	@echo "Remove it manually if no longer needed:"
	@echo "  rm -rf ~/.config/mihomo-cli"

install-completions:
	@echo "Installing shell completions..."
	$(DESTDIR)$(BINDIR)/mihomo-cli completion bash | sudo tee /etc/bash_completion.d/mihomo-cli > /dev/null 2>&1 || \
		echo "Bash: run 'mihomo-cli completion bash | sudo tee /etc/bash_completion.d/mihomo-cli'"
	$(DESTDIR)$(BINDIR)/mihomo-cli completion zsh  | sudo tee /usr/local/share/zsh/site-functions/_mihomo-cli > /dev/null 2>&1 || \
		echo "Zsh:  run 'mihomo-cli completion zsh | sudo tee /usr/local/share/zsh/site-functions/_mihomo-cli'"
	$(DESTDIR)$(BINDIR)/mihomo-cli completion fish | sudo tee /usr/local/share/fish/completions/mihomo-cli.fish > /dev/null 2>&1 || \
		echo "Fish: run 'mihomo-cli completion fish | sudo tee /usr/local/share/fish/completions/mihomo-cli.fish'"
	@echo "Completions installed. Restart your shell or source the completion file."

clean:
	rm -f mihomo-cli mihomo-cli.exe
	rm -f $(EMBED_FILE)
