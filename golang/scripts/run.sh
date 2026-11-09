#!/usr/bin/env bash
# Convenience launcher: build (if needed) and run the hello_world binary.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
cd "$ROOT"
go run ./cmd/app "$@"
