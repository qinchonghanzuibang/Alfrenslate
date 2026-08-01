SHELL := /bin/zsh
GO ?= go
GOCACHE ?= /tmp/alfrenslate-go-cache
BUILD_DIR := build
STAGE_DIR := $(BUILD_DIR)/workflow
ARTIFACT := Alfrenslate.alfredworkflow

.PHONY: format test vet build icon package verify clean

format:
	@GOCACHE=$(GOCACHE) $(GO) fmt ./...

test:
	@GOCACHE=$(GOCACHE) $(GO) test ./...

vet:
	@GOCACHE=$(GOCACHE) $(GO) vet ./...

build:
	@mkdir -p $(BUILD_DIR)
	@GOCACHE=$(GOCACHE) CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags='-s -w' -o $(BUILD_DIR)/alfrenslate-arm64 ./cmd/alfrenslate
	@GOCACHE=$(GOCACHE) CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags='-s -w' -o $(BUILD_DIR)/alfrenslate-amd64 ./cmd/alfrenslate

icon:
	@mkdir -p workflow
	@if command -v rsvg-convert >/dev/null; then \
	  rsvg-convert -w 512 -h 512 assets/icon.svg -o workflow/icon.png; \
	else \
	  print -u2 -- "rsvg-convert is required to regenerate workflow/icon.png"; exit 1; \
	fi

package: build icon
	@rm -rf -- $(STAGE_DIR)
	@mkdir -p $(STAGE_DIR)/bin
	@cp workflow/info.plist workflow/run.sh workflow/paste.sh workflow/manage-action.sh workflow/icon.png $(STAGE_DIR)/
	@cp $(BUILD_DIR)/alfrenslate-arm64 $(BUILD_DIR)/alfrenslate-amd64 $(STAGE_DIR)/bin/
	@chmod 755 $(STAGE_DIR)/*.sh $(STAGE_DIR)/bin/*
	@rm -f -- $(ARTIFACT) $(ARTIFACT).sha256
	@cd $(STAGE_DIR) && /usr/bin/zip -q -X -r ../../$(ARTIFACT) .
	@/usr/bin/shasum -a 256 $(ARTIFACT) > $(ARTIFACT).sha256

verify: format test vet package
	@/usr/bin/plutil -lint workflow/info.plist
	@version=$$(build/workflow/run.sh version); \
	  [[ "$$version" == "1.0.0" ]]; \
	  [[ "$$version" == "$$("/usr/libexec/PlistBuddy" -c 'Print :version' workflow/info.plist)" ]]; \
	  grep -q "## \[$$version\]" CHANGELOG.md
	@scripts/secret-scan.sh
	@scripts/verify-package.sh
	@if command -v shellcheck >/dev/null; then shellcheck workflow/*.sh scripts/*.sh; else print -- "shellcheck not installed; skipped locally."; fi

clean:
	@rm -rf -- $(BUILD_DIR) $(ARTIFACT) $(ARTIFACT).sha256 workflow/icon.png
