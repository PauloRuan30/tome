.PHONY: build test clean install

BINARY_NAME=tome
INSTALL_DIR=/usr/local/bin

build:
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/tome

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	go clean

install: build
	install -Dm755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)