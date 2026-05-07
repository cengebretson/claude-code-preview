BINARY := claude-code-preview
DEMO_DIR := /tmp/claude-preview-demo
SNAP_DIR := /tmp/claude-snapshots-demo

.PHONY: build test install clean fmt preview clean-preview debug-preview release

build:
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o $(BINARY) .

test:
	go test ./...

install: build
	cp $(BINARY) ~/.local/bin/$(BINARY)

clean:
	rm -f $(BINARY)

release:
	@read -p "Tag version (e.g. v0.1.0): " tag && \
	git tag $$tag && \
	git push origin $$tag

fmt:
	go fmt ./...

preview: build
	@mkdir -p $(DEMO_DIR) $(SNAP_DIR)
	@# Snapshot: before state
	@printf 'package main\n\nfunc greet(name string) string {\n\treturn "Hello, " + name\n}\n' \
		> $(SNAP_DIR)/_tmp_claude-preview-demo_hello.go
	@printf 'package main\n\nvar settings = map[string]string{\n\t"host": "localhost",\n}\n' \
		> $(SNAP_DIR)/_tmp_claude-preview-demo_config.go
	@# Current: after state
	@printf 'package main\n\nimport "fmt"\n\nfunc greet(name string) string {\n\tgreeting := fmt.Sprintf("Hello, %%s!", name)\n\treturn greeting\n}\n\nfunc farewell(name string) string {\n\treturn "Goodbye, " + name\n}\n' \
		> $(DEMO_DIR)/hello.go
	@printf 'package main\n\nvar settings = map[string]string{\n\t"host":    "localhost",\n\t"port":    "8080",\n\t"debug":   "true",\n\t"timeout": "30s",\n}\n' \
		> $(DEMO_DIR)/config.go
	@# Write signal files so TUI picks them up immediately
	@printf '$(DEMO_DIR)/hello.go\n$(DEMO_DIR)/config.go\n' > /tmp/claude-preview-signal
	@printf 'demo' > /tmp/claude-preview-signal.session
	@./$(BINARY)

debug-preview:
	@echo "=== snapshot files ==="
	@ls -la $(SNAP_DIR)/ 2>/dev/null || echo "  SNAP_DIR missing"
	@echo ""
	@echo "=== current files ==="
	@ls -la $(DEMO_DIR)/ 2>/dev/null || echo "  DEMO_DIR missing"
	@echo ""
	@echo "=== signal file ==="
	@cat /tmp/claude-preview-signal 2>/dev/null || echo "  signal file missing"
	@echo ""
	@echo "=== git diff | delta (hello.go) ==="
	@git diff --no-index \
		$(SNAP_DIR)/_tmp_claude-preview-demo_hello.go \
		$(DEMO_DIR)/hello.go 2>/dev/null \
		| delta --file-style omit --hunk-header-style omit; true
	@echo ""
	@echo "=== raw git diff only (hello.go) ==="
	@git diff --no-index \
		$(SNAP_DIR)/_tmp_claude-preview-demo_hello.go \
		$(DEMO_DIR)/hello.go; true

clean-preview:
	@rm -rf $(DEMO_DIR) $(SNAP_DIR)
	@rm -f /tmp/claude-preview-signal /tmp/claude-preview-signal.session
