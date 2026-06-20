PAN302_PLUGIN_CLI ?= ../../bin/pan302-plugin
VERSION := 1.0.0
DIST := dist
PACKAGE_DIR := $(DIST)/package
PACKAGE := $(DIST)/strm-example-$(VERSION).panplugin

.PHONY: build validate clean

build: clean
	cargo build --target wasm32-unknown-unknown --release
	mkdir -p $(PACKAGE_DIR)/ui
	cp target/wasm32-unknown-unknown/release/strm_example.wasm $(PACKAGE_DIR)/plugin.wasm
	cp manifest.json default-config.json config.schema.json $(PACKAGE_DIR)/
	cp ui/manifest.json ui/index.js ui/index.css $(PACKAGE_DIR)/ui/
	$(PAN302_PLUGIN_CLI) build $(PACKAGE_DIR) --output $(PACKAGE)

validate:
	$(PAN302_PLUGIN_CLI) validate $(PACKAGE)

clean:
	rm -rf $(DIST)
