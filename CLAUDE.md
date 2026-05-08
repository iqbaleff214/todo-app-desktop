# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

A minimalist offline-first desktop todo widget built with **Wails v2** (Go + Vue 3 + TypeScript). It floats on the desktop showing today's tasks and can expand into a full window for multi-date management. No cloud, no accounts, no CGO.

All active implementation work is tracked in `TASKS.md`. Check it first to understand what is built, what is next, and the exact constraints for each task. `PRD.md` is the canonical product spec.

## Commands

All Wails commands require the CLI binary at `$(go env GOPATH)/bin/wails`.

```bash
# Development (hot-reload, opens a browser-rendered window)
wails dev

# Production binary (output: build/bin/)
wails build

# Go tests — run from project root
go test ./...
go test -v ./internal/models/...        # single package
go test -v -run TestParseDate ./internal/models/...  # single test
go test -race ./internal/repository/... # with race detector

# Go vet
go vet ./...

# Frontend only (from frontend/)
npm run build     # vue-tsc type-check + vite build
npm run dev       # vite dev server standalone
```

After modifying any exported Go method that is bound to the frontend, regenerate the JS bindings:
```bash
wails generate module
```
This writes `frontend/wailsjs/go/`.

## Architecture

### Request path

```
Vue component
  → Pinia store action
    → wailsjs/go/<Package>/<Method>()   ← auto-generated JS binding
      → Go service method (exported, on App or service struct)
        → Repository interface method
          → SQLite via database/sql ("sqlite" driver from modernc.org/sqlite)
```

There is **no HTTP layer**. Frontend ↔ Go communication is exclusively through Wails bindings. Go emits events to the frontend via `runtime.EventsEmit`.

### Go layer (`internal/`)

| Package | Responsibility |
|---------|---------------|
| `models` | Plain structs (`Task`, `Settings`), validation (`Validate()`, `ParseDate()`), `Default()` |
| `repository` | All I/O: SQLite via `database/sql`, settings JSON file, startup backup |
| `service` | Business logic (not yet implemented — Phase 2) |
| `window` | Window mode switching, tray icon, autostart (not yet implemented — Phase 3) |

**Key repository details:**
- SQLite driver registered by blank import in `internal/repository/driver.go`; driver name is `"sqlite"`
- `repository.Open(dataDir)` applies `PRAGMA journal_mode=WAL` and `PRAGMA foreign_keys=ON`, then runs `CREATE TABLE IF NOT EXISTS` migrations — idempotent on every launch
- `db.SetMaxOpenConns(1)` is intentional: WAL allows concurrent reads but SQLite has a single writer
- Times stored as RFC3339 strings (TEXT column); `done` stored as INTEGER (0/1)
- `SettingsRepository.Save()` is atomic: writes `.tmp` then `os.Rename()`
- `MaybeBackup()` must be called before `repository.Open()` at startup

**Data directory** (resolved by `repository.DataDir()`):
- macOS: `~/Library/Application Support/todo-app/`
- Windows: `%APPDATA%\todo-app\`
- Linux: `~/.config/todo-app/`

### Frontend layer (`frontend/src/`)

Vue 3 Composition API (`<script setup>`), Pinia stores, `@vueuse/core`. **No UI component libraries, no CSS frameworks.** All components are custom. CSS custom properties for theming (`data-theme` attribute on `<html>`).

`frontend/wailsjs/` is **auto-generated** — never edit it by hand.

## Hard constraints (from TASKS.md / PRD.md)

- **No CGO**: use `modernc.org/sqlite`, never `mattn/go-sqlite3`
- **No JSON tags on `Task` struct** — serialisation is the service layer's concern
- **`Settings` struct requires JSON tags** — it is written directly to `settings.json`
- All repository errors must wrap with method context: `fmt.Errorf("task_repo.Create: %w", err)`
- `GetByDate` must always return a non-nil slice
- `Delete` must not error on a missing ID
- `MaybeBackup` logs errors but must not crash the app (caller handles gracefully)
- Date strings are strictly `YYYY-MM-DD`; use `models.ParseDate()` to validate before storing
- Task text: max 500 chars; validate with `task.Validate()` before any persistence call

## Wails-specific notes

- `main.go` embeds `frontend/dist` — a `wails build` triggers `npm run build` automatically
- `wails dev` serves the frontend via Vite's dev server (hot-reload); no embed needed
- Frameless window with `AlwaysOnTop` is the widget mode; title bar drag uses `-webkit-app-region: drag` CSS
- Buttons inside a drag region need `-webkit-app-region: no-drag`
- System tray uses `github.com/wailsapp/wails/v2/pkg/menu.TrayMenu` — no third-party systray package
