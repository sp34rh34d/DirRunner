BINARY  := dirrunner
PKG     := ./cmd/dirrunner
DIST    := dist
VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# OS/ARCH pairs to cross-compile for `make release`.
PLATFORMS := \
	linux/amd64 linux/arm64 linux/386 linux/arm \
	darwin/amd64 darwin/arm64 \
	windows/amd64 windows/arm64 windows/386 \
	freebsd/amd64 openbsd/amd64

.PHONY: all build install test vet fmt tidy clean release $(PLATFORMS)

all: build

## build: compile for the current OS/ARCH into ./$(BINARY)
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

## install: install the binary into $(GOBIN)/$(GOPATH)/bin
install:
	go install -ldflags "$(LDFLAGS)" $(PKG)

## test: run the test suite
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go sources
fmt:
	gofmt -s -w .

## tidy: tidy go.mod
tidy:
	go mod tidy

## release: cross-compile for every platform in $(PLATFORMS) into ./$(DIST)
release: $(PLATFORMS)

# Pattern rule: one target per OS/ARCH pair. Produces
# dist/dirrunner-<os>-<arch>[.exe]
$(PLATFORMS):
	@mkdir -p $(DIST)
	$(eval GOOS := $(word 1,$(subst /, ,$@)))
	$(eval GOARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	@echo "building $(GOOS)/$(GOARCH)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -trimpath -ldflags "$(LDFLAGS)" \
		-o $(DIST)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT) $(PKG)

## clean: remove build artifacts
clean:
	rm -rf $(DIST) $(BINARY)

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
