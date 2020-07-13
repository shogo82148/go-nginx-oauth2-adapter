GOVERSION=$(shell go version)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
VERSION=$(patsubst "%",%,$(lastword $(shell grep 'const Version' main.go)))
ARTIFACTS_DIR=$(CURDIR)/artifacts/$(VERSION)
RELEASE_DIR=$(CURDIR)/release/$(VERSION)
SRC_FILES = $(wildcard *.go cli/go-nginx-oauth2-adapter/*.go provider/*.go)
GITHUB_USERNAME=shogo82148

.PHONY: all test clean

all: build-windows-386 build-windows-amd64 build-linux-386 build-linux-amd64 build-darwin-386 build-darwin-amd64

##### build settings

.PHONY: build build-windows-amd64 build-windows-386 build-linux-amd64 build-linux-386 build-darwin-amd64 build-darwin-386

$(ARTIFACTS_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH):
	@mkdir -p $@

$(ARTIFACTS_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)/go-nginx-oauth2-adapter$(SUFFIX): $(ARTIFACTS_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH) $(SRC_FILES)
	@echo " * Building binary for $(GOOS)/$(GOARCH)..."
	@./run-in-docker.sh go build -o artifacts/$(VERSION)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)/go-nginx-oauth2-adapter$(SUFFIX) cli/go-nginx-oauth2-adapter/main.go

build: $(ARTIFACTS_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)/go-nginx-oauth2-adapter$(SUFFIX)

build-windows-amd64:
	@$(MAKE) build GOOS=windows GOARCH=amd64 SUFFIX=.exe

build-windows-386:
	@$(MAKE) build GOOS=windows GOARCH=386 SUFFIX=.exe

build-linux-amd64:
	@$(MAKE) build GOOS=linux GOARCH=amd64

build-linux-386:
	@$(MAKE) build GOOS=linux GOARCH=386

build-darwin-amd64:
	@$(MAKE) build GOOS=darwin GOARCH=amd64

build-darwin-386:
	@$(MAKE) build GOOS=darwin GOARCH=386

##### release settings

.PHONY: release-windows-amd64 release-windows-386 release-linux-amd64 release-linux-386 release-darwin-amd64 release-darwin-386
.PHONY: release-targz release-zip release-files release-upload

$(RELEASE_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH):
	@mkdir -p $@

release-windows-amd64:
	@$(MAKE) release-zip GOOS=windows GOARCH=amd64 SUFFIX=.exe

release-windows-arm64:
	@$(MAKE) release-zip GOOS=windows GOARCH=arm64 SUFFIX=.exe

release-windows-386:
	@$(MAKE) release-zip GOOS=windows GOARCH=386 SUFFIX=.exe

release-linux-amd64:
	@$(MAKE) release-targz GOOS=linux GOARCH=amd64

release-linux-arm64:
	@$(MAKE) release-targz GOOS=linux GOARCH=arm64

release-linux-386:
	@$(MAKE) release-targz GOOS=linux GOARCH=386

release-darwin-amd64:
	@$(MAKE) release-targz GOOS=darwin GOARCH=amd64

release-darwin-arm64:
	@$(MAKE) release-targz GOOS=darwin GOARCH=arm64

release-darwin-386:
	@$(MAKE) release-targz GOOS=darwin GOARCH=386

release-targz: build $(RELEASE_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)
	@echo " * Creating tar.gz for $(GOOS)/$(GOARCH)"
	tar -czf $(RELEASE_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH).tar.gz -C $(ARTIFACTS_DIR) go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)

release-zip: build $(RELEASE_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)
	@echo " * Creating zip for $(GOOS)/$(GOARCH)"
	cd $(ARTIFACTS_DIR) && zip -9 $(RELEASE_DIR)/go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH).zip go-nginx-oauth2-adapter_$(GOOS)_$(GOARCH)/*

release-files: release-windows-386 release-windows-amd64 release-windows-arm64 release-linux-386 release-linux-amd64 release-linux-arm64 release-darwin-386 release-darwin-amd64 release-darwin-arm64

release-upload: release-files
	ghr -u $(GITHUB_USERNAME) --draft --replace v$(VERSION) $(RELEASE_DIR)

test:
	go test -v -race ./...

clean:
	rm -rf vendor
	rm -rf artifacts
	rm -rf release
	if [ -e .mod ] ; then chmod -R +w .mod/; fi
	rm -rf .mod
