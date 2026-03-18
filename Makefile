.PHONY: build test test-single lint clean install render render-svg render-png

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

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME)

render: render-svg render-png

render-svg: install
	@mkdir -p tests/output
	@for f in tests/input/*.d2; do \
		base=$$(basename "$${f%.d2}"); \
		d2 --layout=octopus "$$f" "tests/output/$${base}.svg"; \
	done

render-png: install
	@mkdir -p tests/png
	@for f in tests/input/*.d2; do \
		base=$$(basename "$${f%.d2}"); \
		d2 --layout=octopus "$$f" "tests/png/$${base}.png"; \
	done
