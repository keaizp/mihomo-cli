.PHONY: build clean download-kernel

MIHOMO_VERSION := v1.18.10
EMBED_DIR := internal/kernel/embedded
EMBED_FILE := $(EMBED_DIR)/mihomo-linux-amd64.gz
MIHOMO_URL := https://github.com/MetaCubeX/mihomo/releases/download/$(MIHOMO_VERSION)/mihomo-linux-amd64-$(MIHOMO_VERSION).gz

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

clean:
	rm -f mihomo-cli mihomo-cli.exe
	rm -f $(EMBED_FILE)
