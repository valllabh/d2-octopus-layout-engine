.PHONY: build test test-single lint clean install run vet

BINARY_NAME=d2plugin-octopus
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/d2plugin-octopus

test:
	go test ./...

test-single:
	go test -run $(RUN) $(PKG)

lint:
	go vet ./...

vet: lint

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME)

run: build
	@echo "Plugin built at $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Usage: d2 --layout=octopus diagram.d2"
