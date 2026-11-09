# Hello World (Go) — build & run shortcuts, mirrors the risk_audit layout.

GO     ?= go
BIN     = bin
APP     = $(BIN)/app
PKG     = ./cmd/app

.PHONY: all build run clean test tidy

all: build

build:
	@mkdir -p $(BIN)
	@$(GO) build -o $(APP) $(PKG)

run:
	@$(GO) run $(PKG)

test:
	@$(GO) test ./...

tidy:
	@$(GO) mod tidy

clean:
	@rm -rf $(BIN)
