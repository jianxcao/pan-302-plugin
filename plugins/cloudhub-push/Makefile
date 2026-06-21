PAN302_PLUGIN_CLI ?= ../../bin/pan302-plugin
VERSION := 1.0.1
DIST := dist
PACKAGE_DIR := $(DIST)/package
PACKAGE := $(DIST)/cloudhub-push-$(VERSION).panplugin

.PHONY: build validate clean

build: clean
	cd ui && pnpm install --no-frozen-lockfile && pnpm run build
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -ldflags="-s -w" -o $(DIST)/plugin.wasm .
	mkdir -p $(PACKAGE_DIR)/ui/out
	cp $(DIST)/plugin.wasm $(PACKAGE_DIR)/plugin.wasm
	cp manifest.json default-config.json config.schema.json $(PACKAGE_DIR)/
	cp ui/manifest.json $(PACKAGE_DIR)/ui/
	cp -r ui/out/* $(PACKAGE_DIR)/ui/out/
	$(PAN302_PLUGIN_CLI) build $(PACKAGE_DIR) --output $(PACKAGE)

validate:
	$(PAN302_PLUGIN_CLI) validate $(PACKAGE)

clean:
	rm -rf $(DIST)
	rm -rf ui/out

