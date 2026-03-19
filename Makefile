.PHONY: build test test-single test-upstream lint clean install render render-svg render-png

BINARY_NAME=d2plugin-octopus
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/d2plugin-octopus

test: test-unit test-upstream

test-unit:
	go test ./...

test-single:
	go test -run $(RUN) $(PKG)

test-upstream: install
	@passed=0; failed=0; total=0; \
	for f in tests/input/d2-upstream/*.d2; do \
		total=$$((total+1)); \
		base=$$(basename "$$f"); \
		if d2 --layout=octopus "$$f" /tmp/octopus-test-out.svg 2>/dev/null; then \
			passed=$$((passed+1)); \
		else \
			failed=$$((failed+1)); \
			echo "FAIL: $$base"; \
		fi; \
	done; \
	echo "D2 upstream: $$passed/$$total passed"; \
	if [ $$failed -gt 0 ]; then exit 1; fi

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
