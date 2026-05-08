# Development Tasks — todo-app

Tracks all implementation work derived from PRD v1.0.
Tasks are ordered by dependency. Complete phases in sequence; tasks within a phase can parallelize.

**Legend:** Each task lists **What**, **Output**, **Constraints**, and **Done when**.

---

## Phase 0 — Project Bootstrap

- [x] **T-001 · Initialize Wails project**

  **What:** Scaffold the Wails v2 project with Go module and Vue 3 + TypeScript template.

  **Output:** Runnable `wails dev` and `wails build` with default hello-world frontend.

  **Constraints:**
  - Use `wails init -n todo-app -t vue-ts`
  - Go module path: `github.com/iqbaleff214/todo-app`
  - Wails version: pin to latest stable v2 release in `go.mod`
  - Node version: 18+ (LTS)

  **Done when:** `wails dev` opens a window; `wails build` produces a binary with no errors on the dev machine.

---

- [x] **T-002 · Set up folder structure**

  **What:** Create all directories and empty placeholder files matching PRD §10.

  **Output:**
  ```
  internal/models/
  internal/repository/
  internal/service/
  internal/window/
  frontend/src/components/
  frontend/src/stores/
  frontend/src/composables/
  frontend/src/styles/
  ```

  **Constraints:**
  - Do not move or rename Wails-generated files (`main.go`, `app.go`, `wails.json`)
  - Add `.gitkeep` in empty dirs so they are tracked

  **Done when:** `git status` shows the full tree; project still compiles.

---

- [x] **T-003 · Add Go dependencies**

  **What:** Add all required Go packages to `go.mod` / `go.sum`.

  **Packages:**
  | Package | Purpose |
  |---------|---------|
  | `modernc.org/sqlite` | Pure-Go SQLite driver (no CGO) |
  | `github.com/google/uuid` | UUID v4 generation |
  | `github.com/getlantern/systray` | System tray (or use Wails built-in tray if v2.8+) |

  **Constraints:**
  - `modernc.org/sqlite` must be chosen over `mattn/go-sqlite3` — CGO must not be required
  - Run `go mod tidy` after adding; commit both `go.mod` and `go.sum`

  **Done when:** `go build ./...` succeeds with no CGO required.

---

- [ ] **T-004 · Add frontend dependencies**

  **What:** Install frontend packages via npm inside `frontend/`.

  **Packages:**
  | Package | Purpose |
  |---------|---------|
  | `pinia` | State management |
  | `@vueuse/core` | Composable utilities (useDark, useEventListener) |

  **Constraints:**
  - No UI component libraries (no Element Plus, no Vuetify, no PrimeVue)
  - No CSS frameworks (no Tailwind, no Bootstrap)
  - Keep `package.json` lean — add only what is listed above

  **Done when:** `npm install` inside `frontend/` exits 0; `vite build` succeeds.

---

## Phase 1 — Data Layer (Go)

- [ ] **T-101 · Define Task model**

  **What:** Implement `internal/models/task.go` with the `Task` struct from PRD §8.1.

  **Output:**
  ```go
  type Task struct {
      ID        string
      Date      string    // "YYYY-MM-DD"
      Text      string
      Done      bool
      Position  int
      CreatedAt time.Time
      UpdatedAt time.Time
  }
  ```

  **Constraints:**
  - `Text` max length is 500 chars — enforce with a `Validate() error` method on the struct
  - Date format must be strictly `YYYY-MM-DD`; add a `ParseDate(s string) (string, error)` helper that rejects anything else
  - No JSON tags on the struct itself — serialization is handled at the service layer

  **Done when:** `go vet ./internal/models/...` passes; unit tests cover `Validate()` with empty text, text > 500 chars, invalid date format.

---

- [ ] **T-102 · Define Settings model**

  **What:** Implement `internal/models/settings.go` with the `Settings` struct from PRD §8.2 plus a `Default()` constructor.

  **Output:**
  ```go
  func Default() Settings {
      return Settings{
          Theme:         "system",
          Opacity:       0.9,
          AutoHide:      false,
          LaunchOnLogin: false,
          AllowEditPast: false,
          WindowX:       -1,  // -1 means "use default position"
          WindowY:       -1,
          WidgetWidth:   280,
          WidgetHeight:  420,
      }
  }
  ```

  **Constraints:**
  - `Theme` must be one of `"light"`, `"dark"`, `"system"` — add `ValidateTheme() error`
  - `Opacity` clamped to `[0.4, 1.0]` — add `ClampOpacity()` method
  - JSON tags required — this struct is serialized to `settings.json`

  **Done when:** Unit tests cover `Default()` values, `ValidateTheme()` rejects unknown themes, `ClampOpacity()` clamps correctly.

---

- [ ] **T-103 · Implement database initialization**

  **What:** Implement `internal/repository/db.go` that opens the SQLite file, runs migrations, and returns a `*sql.DB`.

  **Output:** `func Open(dataDir string) (*sql.DB, error)` that:
  1. Creates `dataDir` if it does not exist (`os.MkdirAll`)
  2. Opens `data.db` in that directory
  3. Runs the schema from PRD §8.3 using `CREATE TABLE IF NOT EXISTS`
  4. Sets `PRAGMA journal_mode=WAL` and `PRAGMA foreign_keys=ON`
  5. Returns the open `*sql.DB`

  **Constraints:**
  - Platform data directory resolved by a helper `DataDir() string` using `os.UserConfigDir()` (covers all three platforms)
  - No ORM — raw `database/sql` only
  - Use `modernc.org/sqlite` driver registered as `"sqlite"`

  **Done when:** Integration test creates a temp dir, calls `Open()`, inserts a row, reads it back, asserts equality.

---

- [ ] **T-104 · Implement TaskRepository**

  **What:** Implement `internal/repository/task_repo.go` with full CRUD and date-query methods.

  **Methods:**
  ```go
  type TaskRepository interface {
      GetByDate(date string) ([]Task, error)
      GetDatesWithTasks() ([]string, error)  // returns distinct dates, descending
      Create(task Task) error
      Update(task Task) error
      Delete(id string) error
      ReorderPositions(ids []string) error   // sets position = index for each id
  }
  ```

  **Constraints:**
  - `GetByDate` returns tasks ordered by `position ASC`, then `done ASC`
  - `GetDatesWithTasks` returns at most 365 dates
  - `Create` generates UUID v4 if `task.ID` is empty; sets `created_at` and `updated_at` to `time.Now().UTC()`
  - `Update` always sets `updated_at` to `time.Now().UTC()`; never changes `created_at`
  - All methods must wrap errors with context: `fmt.Errorf("task_repo.Create: %w", err)`

  **Done when:** Integration tests cover each method; `GetByDate` with no rows returns empty slice (not nil); deleting non-existent ID returns no error.

---

- [ ] **T-105 · Implement SettingsRepository**

  **What:** Implement `internal/repository/settings_repo.go` that reads and writes `settings.json`.

  **Methods:**
  ```go
  type SettingsRepository interface {
      Load() (Settings, error)
      Save(s Settings) error
  }
  ```

  **Constraints:**
  - `Load()` returns `models.Default()` when the file does not exist (first launch)
  - `Save()` writes atomically: write to `settings.json.tmp`, then `os.Rename()` to `settings.json`
  - File path: same `dataDir` as the SQLite file
  - JSON must be pretty-printed (`json.MarshalIndent`)

  **Done when:** Unit tests verify: missing file → defaults returned; saved settings survive a round-trip through `Load()`; concurrent `Save()` calls do not corrupt the file.

---

- [ ] **T-106 · Implement startup backup**

  **What:** Implement `internal/repository/backup.go` with a `MaybeBackup(dataDir string) error` function.

  **Logic:**
  1. If `data.db` does not exist → return nil (nothing to back up)
  2. If `data.db.bak` does not exist → copy `data.db` to `data.db.bak` → return nil
  3. Compare `data.db` mtime to `data.db.bak` mtime
  4. If `data.db` is newer → overwrite `data.db.bak` with a copy of `data.db`
  5. Otherwise → do nothing

  **Constraints:**
  - Copy must be done file-to-file (not `exec.Command("cp")`) for cross-platform safety
  - Must be called once per app startup, before any DB writes
  - Log (but do not crash on) backup errors — backup failure must not block the app

  **Done when:** Unit test: creates a fake `data.db`, calls `MaybeBackup`, asserts `.bak` exists and has identical content; second call with unchanged file does not re-copy.

---

## Phase 2 — Service Layer (Go)

- [ ] **T-201 · Implement TaskService**

  **What:** Implement `internal/service/task_service.go` as the business logic layer wrapping `TaskRepository`.

  **Methods exposed to Wails (public, exported):**
  ```go
  func (s *TaskService) GetTasksForDate(date string) ([]Task, error)
  func (s *TaskService) GetTodayTasks() ([]Task, error)
  func (s *TaskService) GetDatesWithTasks() ([]string, error)
  func (s *TaskService) AddTask(date, text string) (Task, error)
  func (s *TaskService) ToggleDone(id string) (Task, error)
  func (s *TaskService) UpdateText(id, text string) (Task, error)
  func (s *TaskService) DeleteTask(id string) error
  func (s *TaskService) ReorderTasks(ids []string) error
  ```

  **Constraints:**
  - `AddTask`: trims whitespace, rejects empty text, rejects text > 500 chars, sets `Position` to `len(existingTasks)` for the date
  - `ToggleDone`: fetches current task, flips `Done`, calls `repo.Update`
  - `GetTodayTasks`: calls `GetTasksForDate(time.Now().Format("2006-01-02"))`
  - All methods must return typed errors the frontend can distinguish (define sentinel errors: `ErrEmptyText`, `ErrTextTooLong`, `ErrTaskNotFound`)

  **Done when:** Unit tests cover all methods; `AddTask` with empty text returns `ErrEmptyText`; `ToggleDone` on missing ID returns `ErrTaskNotFound`.

---

- [ ] **T-202 · Implement SettingsService**

  **What:** Implement `internal/service/settings_service.go` wrapping `SettingsRepository`.

  **Methods exposed to Wails:**
  ```go
  func (s *SettingsService) GetSettings() (Settings, error)
  func (s *SettingsService) SaveSettings(settings Settings) error
  func (s *SettingsService) ResetWindowPosition() error
  ```

  **Constraints:**
  - `SaveSettings`: call `settings.ValidateTheme()` and `settings.ClampOpacity()` before writing
  - `ResetWindowPosition`: sets `WindowX = -1`, `WindowY = -1`, saves
  - Cache settings in memory after first load; invalidate cache on `SaveSettings`
  - Thread-safe: use `sync.RWMutex` around cache reads/writes

  **Done when:** Unit tests verify validation is called; cached read does not hit disk on second `GetSettings()`; concurrent reads are safe.

---

- [ ] **T-203 · Wire services into Wails App struct**

  **What:** Update `app.go` to initialize all services in `startup()` and expose them.

  **Output:**
  ```go
  type App struct {
      ctx             context.Context
      taskService     *service.TaskService
      settingsService *service.SettingsService
      windowManager   *window.Manager
  }
  ```
  `startup()` must:
  1. Resolve `dataDir`
  2. Call `backup.MaybeBackup(dataDir)`
  3. Open DB
  4. Instantiate repos → services
  5. Instantiate `WindowManager`

  **Constraints:**
  - All `TaskService` and `SettingsService` methods must be registered in `main.go` via `wails.Run` `Bind` option
  - Run `wails generate module` after binding to regenerate `wailsjs/`
  - Errors during startup (DB open failure) must surface as a fatal dialog, not a silent crash

  **Done when:** `wails dev` starts; browser devtools show the bound Go methods callable from JS console.

---

## Phase 3 — Window Manager (Go)

- [ ] **T-301 · Implement WindowManager — widget mode**

  **What:** Implement `internal/window/manager.go` to configure and control the Wails window in widget (floating) mode.

  **Behavior on startup:**
  1. Read `Settings.WindowX/Y` — if `-1`, compute center-right of primary screen
  2. Validate position is within any connected screen bounds; snap to nearest screen edge if off-canvas
  3. Set window frameless, `AlwaysOnTop: true`, transparent background, no resize
  4. Restore `WindowX/Y` from settings

  **Methods:**
  ```go
  func (m *Manager) SetWidgetMode()
  func (m *Manager) SetExpandedMode()
  func (m *Manager) SavePosition()          // called on window move
  func (m *Manager) ToggleVisibility()
  ```

  **Constraints:**
  - Use `runtime.WindowSetPosition`, `runtime.WindowSetSize`, `runtime.WindowSetAlwaysOnTop` from Wails runtime
  - Widget default size: `280 × 420` px
  - Expanded min size: `420 × 500` px; max: unconstrained
  - `SavePosition()` must debounce writes — only persist after 500 ms idle (avoid thrashing on drag)

  **Done when:** App launches in widget mode at correct position; dragging updates stored position; re-launch restores position.

---

- [ ] **T-302 · Implement system tray**

  **What:** Add a system tray icon with a context menu.

  **Tray menu items:**
  - Show / Hide (toggles widget visibility)
  - Expand (opens expanded window mode)
  - Separator
  - Quit

  **Left-click:** toggle widget visibility.

  **Constraints:**
  - Use Wails v2 built-in tray API if available (`options.Mac.OnFileOpen` / tray); fall back to `getlantern/systray` package
  - Tray icon must be a 16×16 (Windows/Linux) or 22×22 (macOS) PNG embedded via `//go:embed`
  - Quitting via tray must call `runtime.Quit(ctx)` (not `os.Exit`)
  - Tray must initialize before the main window is shown

  **Done when:** Tray icon appears on all three platforms; all menu items trigger correct behavior; hiding window keeps tray active.

---

- [ ] **T-303 · Implement launch on login**

  **What:** Implement OS-level autostart registration inside `internal/window/autostart.go`.

  **Methods:**
  ```go
  func EnableAutostart(appName, execPath string) error
  func DisableAutostart(appName string) error
  func IsAutostartEnabled(appName string) (bool, error)
  ```

  **Platform implementations:**
  | Platform | Mechanism |
  |----------|-----------|
  | macOS | Write/delete a `.plist` in `~/Library/LaunchAgents/` |
  | Windows | Write/delete registry key `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` |
  | Linux | Write/delete a `.desktop` file in `~/.config/autostart/` |

  **Constraints:**
  - Use `//go:build` tags to separate platform files (`autostart_darwin.go`, `autostart_windows.go`, `autostart_linux.go`)
  - `execPath` must be the absolute path of the running binary (`os.Executable()`)
  - Settings toggle calls `Enable/DisableAutostart` then saves settings

  **Done when:** On each platform, enabling autostart registers the entry; disabling removes it; app launches on login; `IsAutostartEnabled` reflects actual OS state.

---

## Phase 4 — Frontend Foundation

- [ ] **T-401 · Configure Vite and TypeScript**

  **What:** Update `vite.config.ts` and `tsconfig.json` for the project.

  **Constraints:**
  - `vite.config.ts`: add `resolve.alias` for `@` → `./src`
  - `tsconfig.json`: `strict: true`, `target: "ES2022"`, path alias matching Vite
  - Remove Vite default boilerplate from `App.vue` and `style.css`

  **Done when:** `vite build` produces no TypeScript errors; `@/components/Foo.vue` import resolves correctly.

---

- [ ] **T-402 · Set up theme system**

  **What:** Implement `frontend/src/styles/theme.css` and `base.css`.

  **`theme.css`** defines CSS custom properties for both themes:
  ```css
  :root[data-theme="light"] {
    --color-bg: #ffffff;
    --color-surface: #f5f5f5;
    --color-text-primary: #1a1a1a;
    --color-text-secondary: #6b7280;
    --color-accent: #4f46e5;
    --color-done: #9ca3af;
    --color-border: #e5e7eb;
    --shadow-widget: 0 8px 32px rgba(0,0,0,0.12);
    --radius-widget: 10px;
  }
  :root[data-theme="dark"] { /* dark equivalents */ }
  ```

  **`base.css`**: reset, box-sizing, font stack (system-ui), no margin/padding on body, `user-select: none` on root (desktop app).

  **Constraints:**
  - No hardcoded colors anywhere outside `theme.css`
  - `data-theme` attribute set on `<html>` by the settings store watcher
  - "system" theme uses `prefers-color-scheme` media query via JS (`window.matchMedia`)

  **Done when:** Toggling `data-theme` on `<html>` switches all colors; no color defined outside `theme.css`.

---

- [ ] **T-403 · Implement Pinia tasks store**

  **What:** Implement `frontend/src/stores/tasks.ts`.

  **State:**
  ```ts
  interface TasksState {
    tasks: Task[]          // tasks for currently selected date
    selectedDate: string   // "YYYY-MM-DD"
    datesWithTasks: string[]
    loading: boolean
    error: string | null
  }
  ```

  **Actions:**
  ```ts
  loadDate(date: string): Promise<void>
  loadToday(): Promise<void>
  loadDatesWithTasks(): Promise<void>
  addTask(text: string): Promise<void>
  toggleDone(id: string): Promise<void>
  updateText(id: string, text: string): Promise<void>
  deleteTask(id: string): Promise<void>
  reorderTasks(ids: string[]): Promise<void>
  ```

  **Getters:**
  ```ts
  pendingTasks: Task[]    // done === false, sorted by position
  doneTasks: Task[]       // done === true, sorted by position
  doneCount: number
  totalCount: number
  isToday: boolean        // selectedDate === today
  ```

  **Constraints:**
  - All actions call the corresponding Wails-bound Go method from `wailsjs/go/`
  - On error, set `state.error` and emit a global `toast:error` event (do not `throw`)
  - `loadDate` always resets `loading = true` before the call, `false` after

  **Done when:** Store unit tests (vitest) mock Go bindings and verify state transitions for each action.

---

- [ ] **T-404 · Implement Pinia settings store**

  **What:** Implement `frontend/src/stores/settings.ts`.

  **State:** mirrors `Settings` Go struct fields as TypeScript properties.

  **Actions:**
  ```ts
  load(): Promise<void>
  save(patch: Partial<Settings>): Promise<void>
  resetWindowPosition(): Promise<void>
  ```

  **Constraints:**
  - `save(patch)` merges patch into current state, calls Go `SaveSettings`, on success commits to store
  - Watcher on `theme` updates `document.documentElement.dataset.theme` immediately
  - Watcher on `opacity` updates CSS variable `--widget-opacity` immediately (real-time preview)
  - `load()` called once in `App.vue` `onMounted`

  **Done when:** Changing theme in store reflects on `<html>` attribute without page reload.

---

- [ ] **T-405 · Implement useKeyboard composable**

  **What:** Implement `frontend/src/composables/useKeyboard.ts` — a composable that registers global keyboard shortcuts.

  **Interface:**
  ```ts
  function useKeyboard(handlers: {
    onAddTask?: () => void
    onToggleExpand?: () => void
    onOpenSettings?: () => void
    onToday?: () => void
    onPrevDate?: () => void
    onNextDate?: () => void
    onQuit?: () => void
  }): void
  ```

  **Shortcuts to handle** (from PRD §11):
  | Key | Handler |
  |-----|---------|
  | `n` (no modifier, not in input) | `onAddTask` |
  | `Cmd/Ctrl+E` | `onToggleExpand` |
  | `Cmd/Ctrl+,` | `onOpenSettings` |
  | `Cmd/Ctrl+T` | `onToday` |
  | `Cmd/Ctrl+ArrowLeft` | `onPrevDate` |
  | `Cmd/Ctrl+ArrowRight` | `onNextDate` |
  | `Cmd/Ctrl+Q` | `onQuit` |

  **Constraints:**
  - Use `useEventListener(document, 'keydown', ...)` from `@vueuse/core`
  - Skip handler when `event.target` is an `<input>` or `[contenteditable]` (for `n` key)
  - Composable must be called only once, in `App.vue`
  - Handlers passed as `undefined` are silently skipped

  **Done when:** In `wails dev`, pressing each shortcut triggers the correct action; typing `n` inside a text input does NOT trigger `onAddTask`.

---

- [ ] **T-406 · Implement useDate composable**

  **What:** Implement `frontend/src/composables/useDate.ts` with date utilities used across components.

  **Exports:**
  ```ts
  function today(): string                   // "YYYY-MM-DD"
  function formatDisplay(date: string): string  // "Today", "Yesterday", "Mon May 5"
  function prevDate(date: string): string
  function nextDate(date: string): string
  function isToday(date: string): boolean
  function isFuture(date: string): boolean
  function isPast(date: string): boolean
  ```

  **Constraints:**
  - No external date libraries — use native `Intl.DateTimeFormat` for display formatting
  - All functions are pure (no side effects); export as plain functions, not a composable
  - `formatDisplay` special-cases "Today" and "Yesterday"; all other dates use `"EEE MMM D"` format
  - Locale: use `navigator.language`

  **Done when:** Unit tests (vitest) cover all functions with fixed dates including edge cases (Jan 1, Dec 31, leap day).

---

## Phase 5 — Widget UI

- [ ] **T-501 · Build Widget.vue — root floating panel**

  **What:** Implement `frontend/src/components/Widget.vue` as the root component for the floating widget mode.

  **Layout (top to bottom):**
  1. Title bar (drag handle, app name, count badge, collapse/expand/settings buttons)
  2. TaskList (scrollable, today's tasks only)
  3. TaskInput (always visible at bottom of widget)

  **Behavior:**
  - `mouseenter` → set `idle = false` (full opacity)
  - `mouseleave` → after 2 s, set `idle = true` (fade to `settings.opacity`)
  - Opacity transition: `transition: opacity 400ms ease`
  - Auto-hide controlled by `settings.autoHide`; if `false`, always full opacity
  - Double-click on title bar → call Go `ToggleExpanded()`

  **Constraints:**
  - Title bar must have `style="-webkit-app-region: drag"` for native Wails drag-to-move
  - Buttons inside title bar must have `style="-webkit-app-region: no-drag"` to remain clickable
  - Widget width: fixed `280px`; height: flexible, max `420px`, then scroll
  - No scrollbar visible in title bar area

  **Done when:** Widget renders today's tasks; dragging title bar moves the window; idle fade works; count badge shows correct numbers.

---

- [ ] **T-502 · Build TaskList.vue**

  **What:** Implement `frontend/src/components/TaskList.vue` — renders pending tasks then done tasks.

  **Props:**
  ```ts
  props: {
    tasks: Task[]
    readonly: boolean   // true for past dates when AllowEditPast = false
  }
  ```

  **Layout:**
  - Pending tasks section (no header in widget mode; "To Do" label in expanded mode)
  - Done tasks section below, with `"Completed (N)"` count label, collapsible
  - Empty state: slot or default `<EmptyState />` component

  **Drag-to-reorder:**
  - Use native HTML5 drag-and-drop (`draggable`, `dragover`, `drop` events)
  - Only allow reorder within the pending section; done tasks cannot be reordered
  - On drop, compute new `ids` order, call `tasksStore.reorderTasks(ids)`
  - Visual: dragged item gets `opacity: 0.4`; drop target gets `border-top: 2px solid var(--color-accent)`

  **Constraints:**
  - When `readonly = true`, disable drag, hide delete button, disable checkbox click
  - List transitions: use `<TransitionGroup name="task">` with `transform` + `opacity` animation
  - No external DnD library

  **Done when:** Tasks render in correct order; completing a task animates to done section; reorder persists after reload; readonly mode disables all edits.

---

- [ ] **T-503 · Build TaskItem.vue**

  **What:** Implement `frontend/src/components/TaskItem.vue` — single task row.

  **Anatomy:**
  ```
  [ checkbox ] [ task text / inline input ] [ delete btn (hover) ]
  ```

  **States:**
  | State | Trigger | Display |
  |-------|---------|---------|
  | Default | — | Checkbox + text |
  | Editing | Click on text | Text becomes `<input>`, auto-focused |
  | Done | `task.done = true` | Strikethrough + `--color-done` text |
  | Focused | Arrow-key nav | Highlight background |

  **Events emitted:**
  ```ts
  emit('toggle', task.id)
  emit('edit', task.id, newText)
  emit('delete', task.id)
  ```

  **Constraints:**
  - Checkbox animation: CSS scale + color fill transition on check, 150 ms
  - `Space` key when row is focused (not editing) → emit `toggle`
  - `Delete`/`Backspace` key when row is focused (not editing) → emit `delete`
  - Inline edit: `Enter` or `blur` → emit `edit` with new text (if changed); `Escape` → revert text, exit edit mode
  - Delete button: visible only on `hover` via CSS; `opacity: 0` → `opacity: 1` transition

  **Done when:** All states render correctly; keyboard interaction works; edit revert on Escape works; no edit event fired if text unchanged.

---

- [ ] **T-504 · Build TaskInput.vue**

  **What:** Implement `frontend/src/components/TaskInput.vue` — the "add task" input field.

  **Behavior:**
  - Single `<input type="text">` with placeholder "Add a task…"
  - `Enter` → calls `tasksStore.addTask(text)`, clears input on success
  - `Escape` → clears and blurs input
  - Focus triggered externally via `expose({ focus })` when `N` shortcut is pressed
  - Shows character count `N/500` when text length > 400

  **Constraints:**
  - `maxlength="500"` on the input element
  - Disabled when `readonly = true` (past date without `AllowEditPast`)
  - No submit button — keyboard-only add
  - While `tasksStore.loading`, input is disabled and shows a spinner cursor

  **Done when:** Adding a task via Enter appends it to the list and clears the input; pressing N in Widget focuses this input; 500-char limit enforced.

---

- [ ] **T-505 · Build count badge and empty state**

  **What:** Two small sub-components used inside `Widget.vue`.

  **CountBadge.vue:**
  - Displays `"N / M done"` where N = done count, M = total
  - If all done and total > 0: display `"All done ✓"` with accent color
  - If total = 0: hidden

  **EmptyState.vue:**
  - Shown when task list is empty
  - Widget context: SVG illustration + "Nothing for today." + "Press N to add a task." hint
  - Expanded context: just "No tasks on this day." text (prop `context: 'widget' | 'expanded'`)

  **Done when:** Badge updates immediately after toggle; empty state switches based on context prop.

---

## Phase 6 — Expanded Window UI

- [ ] **T-601 · Build expanded window layout**

  **What:** Implement the two-column expanded layout in `App.vue` (or a dedicated `ExpandedView.vue`).

  **Layout:**
  ```
  ┌──────────────┬──────────────────────────────┐
  │  DateSidebar │  TaskList (selected date)    │
  │  (220px)     │  + TaskInput                 │
  └──────────────┴──────────────────────────────┘
  ```

  **Constraints:**
  - Sidebar fixed width `220px`; main area fills remaining space
  - Minimum window size enforced: `420 × 500` px (set via Wails `MinWidth`/`MinHeight`)
  - Expanded mode has a standard window frame (not frameless)
  - Header bar in expanded mode: app name + "Close to widget" button + settings icon

  **Done when:** Expanded window renders both columns; resizing works above minimum; clicking "Close to widget" returns to widget mode.

---

- [ ] **T-602 · Build DateSidebar.vue**

  **What:** Implement `frontend/src/components/DateSidebar.vue`.

  **Content (top to bottom):**
  1. "Today" button — always pinned at top; highlighted if today is selected
  2. "Yesterday" item
  3. Scrollable list of older dates that have tasks (from `tasksStore.datesWithTasks`)
  4. "Future" section: show dates ahead of today that have tasks

  **Item display:** `formatDisplay(date)` label + task count chip `(N)`

  **Constraints:**
  - Selected date highlighted with `var(--color-accent)` background
  - Clicking any item calls `tasksStore.loadDate(date)`
  - `Cmd/Ctrl+←` / `Cmd/Ctrl+→` arrows navigate dates; sidebar scrolls to keep selection visible
  - Dates with zero tasks never appear in sidebar (already filtered by `GetDatesWithTasks`)
  - Sidebar is scrollable; title bar is not

  **Done when:** Sidebar lists all dates with tasks; keyboard navigation updates selection and scrolls; Today is always visible at top.

---

- [ ] **T-603 · Wire expanded view task editing**

  **What:** Ensure full task CRUD works in expanded window (not just widget).

  **Features to verify work in expanded context:**
  - Inline edit (click text → edit → Enter/blur)
  - Delete via hover button and keyboard
  - Drag-to-reorder within pending section
  - Read-only enforcement for past dates when `AllowEditPast = false`
  - `TaskInput` add for current selected date (including future dates)

  **Constraints:**
  - `tasksStore.addTask` must pass `tasksStore.selectedDate` as the date, not always today
  - Confirm read-only mode shows a tooltip "Enable editing past dates in Settings" on hover of disabled input

  **Done when:** All task operations in expanded window behave identically to widget, except scoped to selected date.

---

## Phase 7 — Settings Panel

- [ ] **T-701 · Build SettingsPanel.vue**

  **What:** Implement `frontend/src/components/SettingsPanel.vue` — a slide-in panel or modal.

  **Sections:**

  **Appearance**
  - Theme picker: three buttons (Light / Dark / System), active highlighted
  - Opacity slider: range `0.4`–`1.0`, step `0.05`; live preview on widget while dragging

  **Behavior**
  - Toggle: Auto-hide widget when idle
  - Toggle: Allow editing past dates

  **System**
  - Toggle: Start on login (calls `EnableAutostart` / `DisableAutostart` Go method)
  - Button: Reset widget position (calls `settingsService.ResetWindowPosition`)

  **Keyboard Shortcuts Reference**
  - Static table listing all shortcuts from PRD §11

  **Constraints:**
  - Panel opens with `Cmd/Ctrl+,` or settings icon; closes with `Escape`
  - Opens as an overlay on top of widget or expanded window — not a separate OS window
  - All changes auto-save on input (no Save button needed)
  - Opacity slider updates CSS variable immediately for live preview; Go `SaveSettings` called on `mouseup`/`change`

  **Done when:** All toggles persist across restarts; opacity slider shows live preview; autostart toggle works on the current platform.

---

## Phase 8 — UX Polish

- [ ] **T-801 · First-launch experience**

  **What:** Detect first launch and show onboarding tooltip.

  **Logic:**
  - On `App.vue` mount: if `datesWithTasks` is empty AND today has no tasks → first launch
  - Show a tooltip attached to `TaskInput`: "Your tasks for today. Press N to add one."
  - Tooltip auto-dismisses after 5 s or when user starts typing
  - Never show again (store `firstLaunchSeen: true` in `settings.json`)

  **Constraints:**
  - Tooltip is CSS-only positioning (no third-party tooltip library)
  - Add `firstLaunchSeen` boolean to `Settings` model and repo

  **Done when:** Fresh install shows tooltip once; second launch does not show it; tooltip disappears when user types.

---

- [ ] **T-802 · Toast notification system**

  **What:** Implement a lightweight global toast for error messages.

  **Behavior:**
  - `toast:error` event (emitted from stores on failed Go call) triggers a toast
  - Toast appears at bottom of widget/window: `"Could not save. Check disk space."`
  - Auto-dismisses after 4 s; click to dismiss early
  - Max one toast visible at a time (queue if multiple errors)

  **Constraints:**
  - Implement as a single `ToastContainer.vue` in `App.vue`, listening to `mitt` event bus or `provide/inject`
  - No third-party toast library
  - Toast does not block interaction with the rest of the UI

  **Done when:** Simulating a Go error triggers toast; toast disappears after 4 s; multiple errors queue correctly.

---

- [ ] **T-803 · Corrupt data recovery flow**

  **What:** On DB open failure (corrupt file), show a recovery dialog instead of crashing.

  **Flow:**
  1. Go `startup()` catches DB open error
  2. Emits `runtime.EventsEmit(ctx, "db:corrupt", nil)`
  3. Frontend shows a modal: "Data file could not be opened. [Restore from backup] [Start fresh]"
  4. "Restore from backup": Go renames `data.db` to `data.db.corrupt`, copies `data.db.bak` to `data.db`, restarts DB open
  5. "Start fresh": Go deletes `data.db`, creates new empty DB
  6. Modal closes; app continues normally

  **Constraints:**
  - Modal must appear before any task data is loaded
  - If no backup exists, "Restore from backup" button is disabled with tooltip "No backup available"
  - Log the corruption event to stderr with timestamp

  **Done when:** Replacing `data.db` with garbage triggers the modal; both recovery paths result in a working app.

---

- [ ] **T-804 · Widget auto-hide and idle fade**

  **What:** Implement the idle/active opacity behavior on the widget.

  **Behavior:**
  - On `mouseenter`: immediately set full opacity; cancel any pending idle timer
  - On `mouseleave`: start 2 s timer; on expiry, fade to `settings.opacity`
  - If `settings.autoHide = false`: widget always stays at full opacity regardless of cursor

  **Constraints:**
  - Timer: use `setTimeout`, cancel with `clearTimeout` on re-enter
  - Opacity transition via CSS (`transition: opacity 400ms ease`) — do NOT use JS animation
  - The Go window itself stays at full opacity; only the CSS `opacity` on the root element changes (to preserve hit-testing)

  **Done when:** Cursor leaving widget fades it after 2 s; returning cursor immediately restores full opacity; toggling `autoHide` in settings takes effect without restart.

---

- [ ] **T-805 · Task list transitions**

  **What:** Animate task addition, completion, deletion, and reorder.

  **Animations:**
  | Event | Animation |
  |-------|-----------|
  | Task added | Slide down + fade in, 200 ms |
  | Task deleted | Fade out + collapse height, 200 ms |
  | Task toggled done | Move to done section: fade out position, fade in at bottom, 300 ms |
  | Reorder drop | Smooth reposition via `<TransitionGroup>` |

  **Constraints:**
  - Use Vue `<TransitionGroup name="task">` with `enter-active-class`, `leave-active-class`
  - `leave-active` must use `position: absolute` to avoid layout jump during height collapse
  - Animations must not block keyboard input (no `pointer-events: none` on list during animation)

  **Done when:** All four animations play smoothly at 60fps; no layout jump; no stale items remain after animation completes.

---

## Phase 9 — Cross-Platform Build

- [ ] **T-901 · Create app icons**

  **What:** Produce platform-specific app icons and place them in `build/`.

  **Required files:**
  | File | Platform | Size |
  |------|----------|------|
  | `build/darwin/todo-app.icns` | macOS | Multi-resolution ICNS |
  | `build/windows/icon.ico` | Windows | Multi-resolution ICO (16, 32, 48, 256 px) |
  | `build/linux/todo-app.png` | Linux | 512×512 PNG |
  | `build/tray-icon.png` | All (tray) | 22×22 PNG |

  **Constraints:**
  - Icon design: simple checkmark or sticky-note motif, single color on transparent background
  - macOS ICNS generated from a 1024×1024 master PNG using `iconutil` or equivalent
  - Windows ICO generated using ImageMagick: `convert icon.png -define icon:auto-resize icon.ico`

  **Done when:** `wails build` on each platform produces a binary with the correct icon visible in Finder/Explorer/file manager.

---

- [ ] **T-902 · Configure wails.json for production build**

  **What:** Update `wails.json` with correct metadata for distribution.

  **Fields to set:**
  ```json
  {
    "name": "todo-app",
    "outputfilename": "todo-app",
    "info": {
      "companyName": "todo-app",
      "productName": "todo-app",
      "productVersion": "1.0.0",
      "copyright": "MIT",
      "comments": "Minimalist floating todo widget"
    }
  }
  ```

  **Constraints:**
  - macOS: set `NSHighResolutionCapable = true` in `Info.plist`
  - Windows: set `requestedExecutionLevel = asInvoker` (no UAC prompt)
  - Linux: ensure desktop file is included in build output

  **Done when:** Built binary shows correct name and version in OS app info dialogs.

---

- [ ] **T-903 · Verify cross-platform builds**

  **What:** Build and smoke-test the binary on all three target platforms.

  **Checklist per platform:**
  - [ ] `wails build` exits 0
  - [ ] Binary launches; widget appears on screen
  - [ ] App icon shows correctly
  - [ ] System tray icon appears
  - [ ] Add a task; verify it persists after quit and relaunch
  - [ ] Drag widget; verify position persists after quit and relaunch
  - [ ] Theme switching works
  - [ ] Autostart toggle registers/unregisters correctly

  **Constraints:**
  - macOS build: sign with ad-hoc signature for local testing (`codesign --sign -`)
  - Windows build: test on Windows 10 and 11
  - Linux build: test on Ubuntu 22.04 with GNOME and on a tiling WM

  **Done when:** All checklist items pass on all three platforms.

---

## Phase 10 — MVP Verification

- [ ] **T-1001 · MVP acceptance checklist**

  Verify every item from PRD §12 passes end-to-end before tagging v1.0.0.

  - [ ] Floating widget renders, stays on top, repositions, persists position across restarts
  - [ ] Add / complete / delete tasks for today — all operations reflected immediately in UI
  - [ ] Tasks persist across restarts (SQLite) — verified by quitting and relaunching
  - [ ] Expanded window opens with date sidebar; date navigation works via keyboard and sidebar click
  - [ ] Light / dark / system theme — all three modes render correctly; system mode responds to OS theme change
  - [ ] System tray icon visible on all platforms; show/hide/quit all function
  - [ ] Start on login — enable in settings, reboot, verify app launches automatically
  - [ ] Single rolling backup written on startup when DB has changed
  - [ ] All 11 keyboard shortcuts from PRD §11 trigger correct actions
  - [ ] `wails build` succeeds on macOS, Windows, Linux with no errors or warnings

  **Done when:** Every checkbox above is checked. This is the v1.0.0 ship condition.

---

## Backlog (Post-MVP)

These tasks are defined but not scheduled. Pick up after v1.0.0 ships.

- [ ] **B-001** — Natural language date parsing in TaskInput ("buy milk tomorrow")
- [ ] **B-002** — Import / export tasks as JSON
- [ ] **B-003** — Statistics view: tasks completed per day, 30-day streak
- [ ] **B-004** — Tags / labels with color-coded filtering
- [ ] **B-005** — Recurring tasks (daily, weekdays, weekly)
- [ ] **B-006** — OS native notifications for due-time reminders
- [ ] **B-007** — iCloud Drive / Dropbox folder sync (file-based, no server)
- [ ] **B-008** — Multiple named lists beyond date-based grouping
- [ ] **B-009** — Subtasks (one level deep)
- [ ] **B-010** — Auto-updater via Wails update mechanism or GitHub Releases check

---

## Task Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 0 | T-001 – T-004 | Project bootstrap |
| 1 | T-101 – T-106 | Data layer (Go) |
| 2 | T-201 – T-203 | Service layer (Go) |
| 3 | T-301 – T-303 | Window manager (Go) |
| 4 | T-401 – T-406 | Frontend foundation |
| 5 | T-501 – T-505 | Widget UI |
| 6 | T-601 – T-603 | Expanded window UI |
| 7 | T-701 | Settings panel |
| 8 | T-801 – T-805 | UX polish |
| 9 | T-901 – T-903 | Cross-platform build |
| 10 | T-1001 | MVP acceptance |
| — | B-001 – B-010 | Post-MVP backlog |

**Total MVP tasks: 37** · **Backlog items: 10**
