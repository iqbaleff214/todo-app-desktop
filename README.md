# todo-app

A minimalist, offline-first desktop todo widget. Floats on your desktop showing today's tasks. Expand it when you need more.

> **Your tasks. Today. On your desktop.**

---

## Features

- **Floating widget** — always-on-top panel, drag to reposition, persists across restarts
- **Today by default** — opens on the current date, zero configuration required
- **Expand to browse** — date sidebar for viewing and managing tasks on any date
- **Keyboard-first** — add, complete, delete, and navigate without touching the mouse
- **Offline, always** — no accounts, no sync, no internet; all data stays on your machine
- **Light / dark / system theme** — follows your OS or set it manually
- **Start on login** — optional autostart on macOS, Windows, and Linux
- **Auto-backup** — rolling `data.db.bak` written on every launch when the database has changed

---

## Screenshots

| Widget (default) | Expanded view | Settings |
|------------------|---------------|----------|
| _coming soon_ | _coming soon_ | _coming soon_ |

---

## Tech stack

| Layer | Technology |
|-------|-----------|
| App framework | [Wails v2](https://wails.io) v2.12.0 |
| Backend | Go 1.22+ |
| Frontend | Vue 3.x + TypeScript |
| Bundler | Vite 3.x |
| State | Pinia 3.x |
| Utilities | @vueuse/core 14.x |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |

---

## Requirements

- **Go** 1.22 or later
- **Node.js** 18 or later (with npm)
- **Wails CLI** v2 — install once:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify your environment:

```bash
wails doctor
```

---

## Development

Clone and start the hot-reload dev server:

```bash
git clone https://github.com/iqbaleff214/todo-app.git
cd todo-app
wails dev
```

`wails dev` installs frontend dependencies, starts the Vite dev server, and opens the app window. The Go backend and Vue frontend both hot-reload on save.

### Frontend only

If you need to work on the frontend in isolation:

```bash
cd frontend
npm install
npm run dev   # Vite dev server at http://localhost:5173
```

Note: Go bindings (`wailsjs/`) are stubs in this mode — backend calls will be no-ops.

---

## Build

```bash
# Production binary for the current platform
wails build

# Output locations:
#   macOS   → build/bin/todo-app.app
#   Windows → build/bin/todo-app.exe
#   Linux   → build/bin/todo-app
```

The build command runs `npm run build` (TypeScript check + Vite bundle) automatically before compiling Go.

### Platform notes

| Platform | Notes |
|----------|-------|
| macOS | Ad-hoc signed automatically by Wails. For distribution, sign with your Apple Developer certificate. |
| Windows | No UAC prompt — runs as the current user (`asInvoker`). |
| Linux | Requires `libgtk-3` and `libwebkit2gtk-4.0` on the target machine. |

---

## Testing

```bash
# All packages
go test ./...

# Single package with verbose output
go test -v ./internal/models/...
go test -v ./internal/repository/...

# Single test
go test -v -run TestParseDate ./internal/models/...

# With race detector
go test -race ./...
```

---

## Project structure

```
todo-app/
├── main.go                   # Wails entry point; embeds frontend/dist
├── app.go                    # App struct and lifecycle hooks
├── wails.json                # Wails project config
├── internal/
│   ├── models/               # Task, Settings structs; validation helpers
│   ├── repository/           # SQLite (task_repo), JSON settings, backup
│   ├── service/              # Business logic — wired to Wails bindings
│   └── window/               # Widget ↔ expanded mode; tray icon; autostart
└── frontend/
    ├── src/
    │   ├── components/       # Widget, TaskList, TaskItem, DateSidebar, …
    │   ├── stores/           # Pinia: tasks.ts, settings.ts
    │   ├── composables/      # useKeyboard.ts, useDate.ts
    │   └── styles/           # base.css, theme.css (CSS custom properties)
    └── wailsjs/              # Auto-generated Go→JS bindings (do not edit)
```

---

## Data storage

All data is stored locally. No files leave your machine.

| Platform | Location |
|----------|----------|
| macOS | `~/Library/Application Support/todo-app/` |
| Windows | `%APPDATA%\todo-app\` |
| Linux | `~/.config/todo-app/` |

**Files:**

| File | Purpose |
|------|---------|
| `data.db` | SQLite database — all tasks |
| `data.db.bak` | Rolling backup written at startup when the DB has changed |
| `settings.json` | User preferences (theme, opacity, window position, …) |

To reset the app to factory state, delete the data directory.

---

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `N` | Add new task (widget focused) |
| `Enter` | Save task |
| `Escape` | Cancel edit / close settings |
| `Space` | Toggle selected task done |
| `Delete` / `Backspace` | Delete selected task |
| `↑` / `↓` | Navigate task list |
| `Cmd/Ctrl+E` | Toggle expanded window |
| `Cmd/Ctrl+,` | Open settings |
| `Cmd/Ctrl+←` / `→` | Navigate dates (expanded mode) |
| `Cmd/Ctrl+T` | Jump to today |
| `Cmd/Ctrl+Q` | Quit |

---

## Contributing

1. Fork the repository and create a branch from `main`.
2. Check `TASKS.md` for the implementation backlog and pick an unstarted task — or open an issue first for anything outside the existing scope.
3. Follow the conventions already in the codebase:
   - Go: `internal/` packages only; no ORM; raw `database/sql`; errors wrapped with `fmt.Errorf("pkg.Method: %w", err)`
   - Frontend: no UI component libraries; no CSS frameworks; custom components only
   - SQLite: `modernc.org/sqlite` only — no CGO dependencies
4. Add tests alongside your code. Repository-layer tests are integration tests using `t.TempDir()` — no mocks.
5. Run `go test -race ./...` and `npm run build` before opening a pull request.

---

## License

MIT
