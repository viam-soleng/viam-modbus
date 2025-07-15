BIN_PATH=bin
BIN_NAME=viam-modbus
ENTRY_POINT=main.go

BIN=$(BIN_PATH)/$(BIN_NAME)
PACKAGE_NAME=$(BIN_NAME).tar.gz

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
	tar -czf $(PACKAGE_NAME) $(BIN)

