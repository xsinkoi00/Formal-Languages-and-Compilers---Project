# hello_world (Go)

A minimal "Hello, world!" program laid out in the same style as
`C:\development\risk_audit\risk_audit` — the standard Go project
structure (`cmd/`, `internal/`, `configs/`, `scripts/`, `bin/`,
`docs/`).

## Layout

```
golang/
├── cmd/
│   └── app/
│       └── main.go           # program entry point
├── internal/
│   └── greeter/
│       └── greeter.go        # module-private library
├── configs/
│   └── config.yaml           # placeholder config (not read yet)
├── scripts/
│   └── run.sh                # convenience launcher
├── docs/                     # project docs (empty)
├── bin/                      # compiled binaries (gitignored)
├── go.mod
├── Makefile
└── README.md
```

## Build and run

```sh
make run       # go run ./cmd/app
make build     # produces bin/app (or bin/app.exe on Windows)
./bin/app
```

Expected output: `Hello, world!`

## Why this layout

- **`cmd/app/`** — one directory per binary. The package is `main` and the
  file is tiny; all real logic lives under `internal/`.
- **`internal/greeter/`** — the Go toolchain forbids packages under
  `internal/` from being imported by anything outside this module, so
  this is where real code goes.
- **`configs/`, `scripts/`, `docs/`, `bin/`** — mirror the directories
  used by larger Go projects (including `risk_audit`) so this hello world
  can grow into a real service without reshuffling.
