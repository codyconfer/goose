APP := goose
MODULE := $(shell go list -m)
BASE_VERSION ?= 0.1.0
VERSION ?= $(shell BASE_VERSION=$(BASE_VERSION) ./scripts/semver.sh)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
DIRTY ?= $(shell test -n "$$(git status --porcelain 2>/dev/null)" && echo true || echo false)
DIST_DIR := dist
BUILD_DIR := build
APPIMAGE_TEMPLATE_DIR := packaging/linux/appimage
APPIMAGETOOL ?= appimagetool
GOBIN := $(shell go env GOPATH)/bin
GOLANGCI := $(GOBIN)/golangci-lint
GOLANGCI_VERSION := v2.12.2
LDFLAGS := -s -w -X '$(MODULE)/internal/buildinfo.Version=$(VERSION)' -X '$(MODULE)/internal/buildinfo.Commit=$(COMMIT)' -X '$(MODULE)/internal/buildinfo.Date=$(BUILD_DATE)' -X '$(MODULE)/internal/buildinfo.Dirty=$(DIRTY)'

.PHONY: fmt lint test check hooks tools build clean version release release-linux release-darwin release-windows checksums appimage-linux-amd64 appimage-linux-arm64 darwin-arm64 windows-amd64 windows-arm64 require-appimagetool require-zip

# Install dev tooling (golangci-lint) into $(GOBIN).
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

# Format all Go code.
fmt:
	gofmt -w .
	@[ -x "$(GOLANGCI)" ] && "$(GOLANGCI)" fmt || true

# Run static analysis.
lint:
	go vet ./...
	"$(GOLANGCI)" run

# Run the test suite.
test:
	go test ./...

# Everything the pre-commit hook runs.
check: fmt lint test

# Install tooling and point git at the tracked hooks directory.
hooks: tools
	git config core.hooksPath .githooks
	chmod +x .githooks/*
	@echo "pre-commit hook enabled (core.hooksPath=.githooks)"

version:
	@echo "$(VERSION)"

build:
	mkdir -p "$(DIST_DIR)"
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o "$(DIST_DIR)/$(APP)" .

release: clean release-linux release-darwin release-windows checksums

release-linux: appimage-linux-amd64 appimage-linux-arm64

release-darwin: darwin-arm64

release-windows: windows-amd64 windows-arm64

define build_binary
	mkdir -p "$(DIST_DIR)" "$(BUILD_DIR)/$(1)"
	CGO_ENABLED=0 GOOS=$(2) GOARCH=$(3) go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BUILD_DIR)/$(1)/$(APP)$(4)" .
endef

define package_appdir
	rm -rf "$(BUILD_DIR)/$(1)/$(APP).AppDir"
	mkdir -p "$(BUILD_DIR)/$(1)/$(APP).AppDir/usr/bin"
	cp "$(BUILD_DIR)/$(1)/$(APP)" "$(BUILD_DIR)/$(1)/$(APP).AppDir/usr/bin/$(APP)"
	cp "$(APPIMAGE_TEMPLATE_DIR)/AppRun" "$(BUILD_DIR)/$(1)/$(APP).AppDir/AppRun"
	cp "$(APPIMAGE_TEMPLATE_DIR)/$(APP).desktop" "$(BUILD_DIR)/$(1)/$(APP).AppDir/$(APP).desktop"
	cp "$(APPIMAGE_TEMPLATE_DIR)/$(APP).svg" "$(BUILD_DIR)/$(1)/$(APP).AppDir/$(APP).svg"
	chmod +x "$(BUILD_DIR)/$(1)/$(APP).AppDir/AppRun" "$(BUILD_DIR)/$(1)/$(APP).AppDir/usr/bin/$(APP)"
endef

appimage-linux-amd64: require-appimagetool
	$(call build_binary,linux-amd64,linux,amd64,)
	$(call package_appdir,linux-amd64)
	ARCH=x86_64 "$(APPIMAGETOOL)" -n "$(BUILD_DIR)/linux-amd64/$(APP).AppDir" "$(DIST_DIR)/$(APP)_$(VERSION)_linux_x86_64.AppImage"

appimage-linux-arm64: require-appimagetool
	$(call build_binary,linux-arm64,linux,arm64,)
	$(call package_appdir,linux-arm64)
	ARCH=aarch64 "$(APPIMAGETOOL)" -n "$(BUILD_DIR)/linux-arm64/$(APP).AppDir" "$(DIST_DIR)/$(APP)_$(VERSION)_linux_aarch64.AppImage"

darwin-arm64:
	$(call build_binary,darwin-arm64,darwin,arm64,)
	tar -C "$(BUILD_DIR)/darwin-arm64" -czf "$(DIST_DIR)/$(APP)_$(VERSION)_darwin_arm64.tar.gz" "$(APP)"

windows-amd64: require-zip
	$(call build_binary,windows-amd64,windows,amd64,.exe)
	cd "$(BUILD_DIR)/windows-amd64" && zip -q -9 "../../$(DIST_DIR)/$(APP)_$(VERSION)_windows_amd64.zip" "$(APP).exe"

windows-arm64: require-zip
	$(call build_binary,windows-arm64,windows,arm64,.exe)
	cd "$(BUILD_DIR)/windows-arm64" && zip -q -9 "../../$(DIST_DIR)/$(APP)_$(VERSION)_windows_arm64.zip" "$(APP).exe"

checksums:
	cd "$(DIST_DIR)" && rm -f SHA256SUMS && sha256sum * > SHA256SUMS

clean:
	rm -rf "$(BUILD_DIR)" "$(DIST_DIR)"

require-appimagetool:
	@command -v "$(APPIMAGETOOL)" >/dev/null 2>&1 || { echo "appimagetool is required for AppImage builds. Install AppImageKit or set APPIMAGETOOL=/path/to/appimagetool."; exit 1; }

require-zip:
	@command -v zip >/dev/null 2>&1 || { echo "zip is required for Windows release archives."; exit 1; }
