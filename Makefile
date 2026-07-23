# Developer tasks for the Verifi CLI.
# `make check` is the pre-commit gate (see docs/adr/0003 and CLAUDE.md).

BINARY := verifi
BIN    := bin/$(BINARY)

.PHONY: check fmt vet test build run clean tidy lint hooks

check: fmt vet test ## fmt, vet, and race-enabled unit tests (run before every commit)

hooks: ## install the git pre-commit hook (.githooks)
	git config core.hooksPath .githooks
	@echo "hooks installed: pre-commit guards main, checks gofmt, blocks em dashes, runs vet and tests"

fmt: ## format and report anything reformatted
	gofmt -l -w .

vet: ## go vet
	go vet ./...

test: ## unit + fixture tests with the race detector
	go test -race ./...

build: ## build the binary into bin/
	@mkdir -p bin
	go build -o $(BIN) .

run: build ## build and run the welcome splash
	./$(BIN)

tidy: ## keep go.mod tidy
	go mod tidy

# Optional: runs only if golangci-lint is installed locally / in CI.
lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

clean:
	rm -rf bin dist
