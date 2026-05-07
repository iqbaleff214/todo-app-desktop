# Product Requirements Document — todo-app

**Version:** 1.0
**Date:** 2026-05-07
**Stack:** Wails v2 · Go · Vue 3 · TypeScript

---

## 1. Product Overview

todo-app is a minimalist, offline-first desktop widget for daily task management. It floats on the desktop as a compact sticky-note panel showing only today's tasks. Users who need to review or manage tasks from other dates can expand it into a full window. No accounts, no sync, no internet required.

**Tagline:** Your tasks. Today. On your desktop.

---

## 2. Goals

| Goal | Metric |
|------|--------|
| Near-instant startup | Cold launch < 300 ms |
| Zero friction for daily use | Add task in < 3 keystrokes |
| Always visible, never intrusive | Floats above windows, click-through when idle |
| Cross-platform parity | Identical UX on macOS, Windows, Linux |
| No data left on servers | 100% local storage |

---

## 3. Non-Goals

- Cloud sync or multi-device support (future)
- Recurring tasks, reminders, or notifications (future)
- Collaboration or sharing
- Rich text, attachments, or subtasks
- Mobile versions

---

## 4. Core Features

### 4.1 Floating Widget (Default State)

- Compact panel, always-on-top
- Shows today's task list only
- Displays task count badge: `3 / 7 done`
- Single-click task to toggle complete
- Inline add: press `N` or click `+` to append new task
- Drag title bar to reposition anywhere on screen
- Right-click title bar → context menu: Expand, Hide, Quit
- Auto-hides to a slim bar when cursor leaves (optional, user-configurable)
- Semi-transparent background, adapts to system light/dark mode

### 4.2 Expanded Window

- Opens from widget via double-click title bar or `Cmd/Ctrl+E`
- Full-height panel with date navigation sidebar
- Sidebar shows: Today, Yesterday, a scrollable list of past dates with tasks
- Main area: task list for selected date
- Tasks grouped by: Pending → Completed
- Add, edit (click to edit inline), delete tasks
- Keyboard-first navigation throughout

### 4.3 Task Management

- Add task: type text → `Enter` to save
- Edit task: click text → inline edit → `Enter` or blur to save
- Complete task: click checkbox or press `Space` when focused
- Delete task: `Delete`/`Backspace` on selected task, or swipe/hover delete icon
- Reorder tasks: drag-and-drop within same date
- Tasks tied to a calendar date (default: today's date when created)

### 4.4 Date Navigation

- Default view: today
- Navigate dates: `Cmd/Ctrl+←` / `Cmd/Ctrl+→` in expanded window
- "Today" button snaps back to current date
- Past dates are read-only by default (user can enable editing in settings)
- Future dates: allowed, for planning ahead

### 4.5 Settings Panel

- Toggle: auto-hide widget when idle
- Toggle: start on login (OS autostart)
- Toggle: allow editing past dates
- Widget opacity: slider 40%–100%
- Widget position: reset to default
- Theme: Light / Dark / System
- Keyboard shortcuts reference

---

## 5. UX Behavior

### 5.1 First Launch

1. Widget appears center-right of screen
2. Tooltip: "Your tasks for today. Press N to add one."
3. Empty state shows placeholder: "Nothing yet. Press N to add a task."

### 5.2 Widget States

| State | Trigger | Appearance |
|-------|---------|------------|
| Active | Cursor over widget | Full opacity, interactive |
| Idle | Cursor away > 2 s | Fade to configured opacity |
| Collapsed | User collapses | Slim title bar only (~28 px tall) |
| Hidden | User hides | System tray icon only |
| Expanded | Double-click / shortcut | Full window, widget hides |

### 5.3 Empty State

- Widget: soft illustration + "Nothing for today." text
- Expanded sidebar date: "No tasks on this day."

### 5.4 Task Completion Feel

- Checkbox animates on check
- Completed tasks move to bottom of list with strikethrough + reduced opacity
- Count badge updates immediately

### 5.5 Error States

- Failed save (disk full, permissions): toast notification "Could not save. Check disk space."
- Corrupt data file: prompt to reset with backup offer

---

## 6. Floating Widget Behavior

- Implemented via Wails frameless window with `AlwaysOnTop: true`
- Transparent background with rounded corners (CSS + platform shadow)
- Drag-to-move: entire title bar is drag handle
- Resize: disabled in widget mode; enabled in expanded mode (min 420×500 px)
- Window position persisted to storage on every move
- Multi-monitor aware: re-validates position on launch, snaps to nearest screen if off-canvas
- System tray icon: left-click toggles widget visibility; right-click → menu

---

## 7. Local Storage

### 7.1 Storage Engine

- Single SQLite file via `modernc.org/sqlite` (pure Go, no CGO)
- File location:
  - macOS: `~/Library/Application Support/todo-app/data.db`
  - Windows: `%APPDATA%\todo-app\data.db`
  - Linux: `~/.local/share/todo-app/data.db`

### 7.2 Settings File

- JSON file alongside the database: `settings.json`
- Written on every settings change

### 7.3 Backup

- On startup, if DB modified since last backup, copy to `data.db.bak`
- Keep single rolling backup only

---

## 8. Data Models

### 8.1 Task

```go
type Task struct {
    ID          string    // UUID v4
    Date        string    // "YYYY-MM-DD"
    Text        string    // max 500 chars
    Done        bool
    Position    int       // sort order within date
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 8.2 Settings

```go
type Settings struct {
    Theme           string  // "light" | "dark" | "system"
    Opacity         float64 // 0.4–1.0
    AutoHide        bool
    LaunchOnLogin   bool
    AllowEditPast   bool
    WindowX         int
    WindowY         int
    WidgetWidth     int
    WidgetHeight    int
}
```

### 8.3 SQLite Schema

```sql
CREATE TABLE tasks (
    id          TEXT PRIMARY KEY,
    date        TEXT NOT NULL,
    text        TEXT NOT NULL,
    done        INTEGER NOT NULL DEFAULT 0,
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE INDEX idx_tasks_date ON tasks(date);
```

---

## 9. Technical Architecture

```
┌─────────────────────────────────────────┐
│              Wails Runtime              │
│  ┌───────────────┐  ┌────────────────┐  │
│  │  Vue 3 + TS   │  │   Go Backend   │  │
│  │  (renderer)   │◄─►│  (app logic)  │  │
│  │               │  │                │  │
│  │ - Widget view │  │ - TaskService  │  │
│  │ - Expanded    │  │ - SettingsSvc  │  │
│  │   view        │  │ - SQLite repo  │  │
│  │ - Settings    │  │ - Window mgr   │  │
│  └───────────────┘  └────────────────┘  │
└─────────────────────────────────────────┘
             │
        SQLite file
```

### 9.1 Go Backend

- `TaskService`: CRUD operations, date queries
- `SettingsService`: read/write settings JSON
- `WindowManager`: switch between widget and expanded modes, tray icon
- All methods exposed to frontend via Wails bindings (`wails generate`)

### 9.2 Vue Frontend

- Vue 3 Composition API + `<script setup>`
- Pinia for state (tasks, settings, UI state)
- No external UI library — custom components only
- CSS custom properties for theming; `prefers-color-scheme` media query as fallback
- Vite for bundling

### 9.3 Communication

- Frontend calls Go via generated Wails bindings (no HTTP, no REST)
- Go emits events to frontend via `runtime.EventsEmit` for reactive updates

---

## 10. Folder Structure

```
todo-app/
├── main.go                  # Wails entry point
├── app.go                   # App struct, lifecycle hooks
├── wails.json
├── build/
│   └── ...                  # Platform icons, manifests
├── internal/
│   ├── models/
│   │   ├── task.go
│   │   └── settings.go
│   ├── repository/
│   │   ├── task_repo.go     # SQLite queries
│   │   └── settings_repo.go
│   ├── service/
│   │   ├── task_service.go
│   │   └── settings_service.go
│   └── window/
│       └── manager.go       # Widget ↔ expanded toggle, tray
└── frontend/
    ├── index.html
    ├── vite.config.ts
    ├── src/
    │   ├── main.ts
    │   ├── App.vue
    │   ├── components/
    │   │   ├── Widget.vue       # Floating widget root
    │   │   ├── TaskItem.vue
    │   │   ├── TaskList.vue
    │   │   ├── TaskInput.vue
    │   │   ├── DateSidebar.vue  # Expanded view only
    │   │   └── SettingsPanel.vue
    │   ├── stores/
    │   │   ├── tasks.ts
    │   │   └── settings.ts
    │   ├── composables/
    │   │   ├── useKeyboard.ts
    │   │   └── useDate.ts
    │   └── styles/
    │       ├── base.css
    │       └── theme.css
    └── wailsjs/              # Auto-generated bindings
        └── go/
```

---

## 11. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `N` | Add new task (widget focused) |
| `Enter` | Save task (edit mode) |
| `Escape` | Cancel edit / close settings |
| `Space` | Toggle selected task done |
| `Delete` / `Backspace` | Delete selected task |
| `↑` / `↓` | Navigate task list |
| `Cmd/Ctrl+E` | Toggle expanded window |
| `Cmd/Ctrl+,` | Open settings |
| `Cmd/Ctrl+←` / `→` | Navigate dates (expanded) |
| `Cmd/Ctrl+T` | Jump to today |
| `Cmd/Ctrl+Q` | Quit app |

---

## 12. MVP Scope

MVP ships when these and only these work correctly:

- [ ] Floating widget renders, stays on top, repositions, persists position
- [ ] Add / complete / delete tasks for today
- [ ] Tasks persist across restarts (SQLite)
- [ ] Expanded window with date sidebar and navigation
- [ ] Light / dark / system theme
- [ ] System tray icon with show/hide/quit
- [ ] Start on login (macOS, Windows, Linux)
- [ ] Single rolling backup of data file
- [ ] All keyboard shortcuts functional
- [ ] Builds on macOS, Windows, Linux via `wails build`

---

## 13. Future Improvements

Ordered by likely value:

1. **Cloud sync** — optional, end-to-end encrypted, via user-provided backend or simple file sync (iCloud Drive, Dropbox folder)
2. **Recurring tasks** — daily/weekly patterns
3. **Due-time reminders** — OS native notifications
4. **Tags / labels** — color-coded categories
5. **Natural language input** — parse "buy milk tomorrow" into date + text
6. **Import/export** — CSV or JSON dump
7. **Subtasks** — one level deep only
8. **Statistics view** — tasks completed per day/week chart
9. **Multiple lists** — named lists beyond date-based grouping
10. **Mobile companion** — iOS/Android app with sync (requires cloud)

---

## 14. Open Questions

| # | Question | Owner | Status |
|---|----------|-------|--------|
| 1 | Widget corner radius: 8 px or 12 px? | Design | Open |
| 2 | Auto-hide delay: 2 s fixed or user-configurable? | PM | Open |
| 3 | Past-date edit: off by default or on? | PM | Open |
| 4 | Drag-to-reorder: within-date only or across dates? | Eng | Open |
| 5 | Wails v2 vs v3 (alpha)? | Eng | Open |
