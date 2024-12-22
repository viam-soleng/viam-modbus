BIN_PATH=bin
BIN_NAME=viam-modbus
ENTRY_POINT=main.go
VERSION_PATH=common/utils.go
PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')/$(shell uname -m | sed 's/aarch64/arm64/')

BIN=$(BIN_PATH)/$(BIN_NAME)
PACKAGE_NAME=$(BIN_NAME).tar.gz

VERSION=$(shell grep 'Version' $(VERSION_PATH) | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
GIT_VERSION=$(shell git describe --tags --abbrev=0 | sed 's/^v//')

ifneq ($(VERSION),$(GIT_VERSION))
$(warning VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match)
endif

default: build

build:
	go mod tidy
	go build -o $(BIN) $(ENTRY_POINT)

test: 
	go test -v -coverprofile=coverage.txt -covermode=atomic ./... -race

clean-package:
	rm -rf $(PACKAGE_NAME)

clean-bin:
	rm -rf $(BIN_PATH)

package: clean-package build
	tar -czf $(PACKAGE_NAME) $(BIN) meta.json

upload: package
	$(info PLATFORM is $(PLATFORM))
	@if [ "$(VERSION)" != "$(GIT_VERSION)" ]; then \
        echo "VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match"; \
        exit 1; \
    fi
	viam module update
	viam module upload --version=$(VERSION) --platform=$(PLATFORM) $(PACKAGE_NAME)
