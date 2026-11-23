# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

@AGENTS.md

## More commands

```sh
make test                                   # go vet + go test ./...
go test ./internal/memory -run TestName -v  # one Go test
gofmt -w <files>                            # required on Go code
make ui                                     # rebuild ui/ into internal/server/ui (commit the output)
make web-test                               # site type check + Playwright against the build
cd web && npm run types:check               # site types only
cd web && npx playwright test e2e/<file>    # one site test
make py-test / make js-test                 # SDK tests; both build bin/woolly and run against a real woolly serve
```

Site dev: `cd web && npm run dev` (http://localhost:5173). Review page dev: `cd ui && npm run dev` proxies to a running `woolly serve`.

## Architecture

Woolly is one Go binary (Go 1.22, standard library only) that stores memories as plain Markdown files under `WOOLLY_HOME`.

- `cmd/woolly/main.go` dispatches subcommands with a `switch` in `run()`; each command lives in its own file in `cmd/woolly` (`check.go`, `why.go`, `connect.go`, ...).
- `internal/memory` is the store. Every write goes through secret scanning (`internal/secrets`) and an instruction-injection check, returning `SensitiveError` / `InjectionError`. `ground.go` defines `Ref` (file, symbol or line range plus a code fingerprint) which is how facts are tied to code and later flagged stale.
- Agents reach the store three ways, all sharing `internal/memory` and `internal/recall`:
  - `internal/mcp`: MCP server (`woolly mcp`) for Claude Code, Cursor, Codex, etc.
  - `internal/server`: local HTTP API (`woolly serve`, `v1.go`) used by the SDKs, plus the embedded review page from `ui/`.
  - `internal/proxy`: chat-request proxy for apps without MCP; injects memories and saves stated facts, streams replies untouched.
- `internal/connect` wires Woolly into each supported agent app; `internal/service` runs it as a background service.
- `ui/` (React) is built into `internal/server/ui` and embedded; `app/` is a Tauri desktop wrapper around the same screens.
- `web/` is a React Router + Fumadocs site deployed on Vercel (root `web`). Docs pages come from top-level `docs/*.mdx` and must be listed in `docs/meta.json`.

## Constraints (from CONTRIBUTING.md)

- Nothing may send data off the user's machine. Memories must stay hand-editable Markdown.
- Changes to how data is stored, deleted or exposed need a test, and an update to `docs/security.mdx` if they change what Woolly protects against.
- After changing what commands print, re-record `docs/demo.gif` with `make demo`.
