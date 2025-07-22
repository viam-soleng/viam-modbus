.PHONY: tidy
ENTRY_POINT=main.go

BIN=$(BIN_PATH)/$(BIN_NAME)
PACKAGE_NAME=$(BIN_NAME).tar.gz


GO_BUILD_ENV :=
GO_BUILD_FLAGS :=
MODULE_BINARY := bin/viam-modbus

$(MODULE_BINARY): tidy Makefile go.mod *.go cmd/module/*.go
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(MODULE_BINARY) cmd/module/cmd.go

tidy:
	go mod tidy

module: module.tar.gz


module.tar.gz: meta.json $(MODULE_BINARY)
	tar czf $@ meta.json $(MODULE_BINARY)
