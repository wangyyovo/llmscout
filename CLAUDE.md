# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**LLM Scout** — a desktop LLM debugging proxy built with Wails v2. Captures and forwards API requests to LLM providers, records full request/response payloads, and provides a visual log viewer with instant replay capabilities.

- **Backend**: Go 1.25 (`go.mod` root, module `github.com/llmscout/llmscout`)
- **Frontend**: Vue 3 + NaiveUI + Vite 5 (`frontend/`)
- **Desktop framework**: Wails v2.12.0

## Commands

### Development

```bash
wails dev          # Run full app in dev mode (hot-reload Go + frontend)
```

From `frontend/` directory:
```bash
npm run dev        # Run Vite dev server standalone (frontend only, on :5173)
npm run build      # Production build of frontend only
npm run preview    # Preview the frontend production build locally
```

### Building

```bash
wails build                          # Production build for current OS/arch
wails build --clean --platform windows/amd64   # Windows AMD64
wails build --clean --platform darwin/universal # macOS universal
wails build --clean --platform darwin/arm64     # macOS Apple Silicon
```

## Architecture

### Backend (Go)

```
internal/
  route/        RouteManager — CRUD + path matching (prefix/exact rules)
  proxy/        ProxyEngine  — HTTP forwarding with SSE streaming support
  log/          LogService   — async channel-based logging + query/filter/paginate
  storage/      SQLite (modernc.org/sqlite, CGO-free)
```

**Database location**: `./data/llmscout.db` (relative to app working directory) with WAL mode.

### Wails Binding Pattern

Go methods on `App` struct in `app.go` are exposed to the frontend via Wails Bind. The `frontend/wailsjs/` directory is auto-generated — never edit it manually.

**Key constraint**: Wails v2 Bind cannot serialize `time.Time`. All timestamp fields exposed to the frontend use `int64` (UnixMilli).

### Key Backend Methods

| Category | Methods |
|----------|---------|
| Route | `ListRoutes`, `AddRoute`, `UpdateRoute`, `DeleteRoute` |
| Proxy | `StartProxy`, `StopProxy`, `ProxyStatus` |
| Log | `QueryLogs`, `GetLog`, `ClearLogs`, `DeleteLogs`, `GetLogRouteNames` |
| Settings | `GetSetting`, `SetSetting`, `GetDbPath` |

### Frontend Structure

```
frontend/src/
  App.vue                    — Root: sidebar layout + NConfigProvider theme
  main.js                    — Bootstrap, naive-ui global registration
  composables/
    useTheme.js              — dark/light/system theme with CSS variables
  views/
    ProxyPanel.vue           — Proxy start/stop, port, stats
    RoutePanel.vue           — Route rules CRUD
    LogViewer.vue            — Log table + detail modal
    SettingsPanel.vue        — Theme toggle, DB path, clear logs
  components/
    LlmMessageViewer.vue     — LLM message parser (OpenAI + Anthropic)
    MarkdownRenderer.vue     — markdown-it rendering
    JsonViewer.vue           — Pretty-print JSON with copy
    HeadersViewer.vue        — Key-value header display with masking
```

### Theme System

CSS variables defined in `App.vue` for both dark/light themes:
`--bg-main`, `--bg-card`, `--bg-code`, `--bg-message`, `--bg-hover`,
`--border-color`, `--text-primary`, `--text-secondary`, `--text-muted`, `--accent`

Theme persistence via `SetSetting('theme', ...)`.

### LLM Protocol Support

The `LlmMessageViewer` parses both **OpenAI** and **Anthropic (Claude)** formats:
- Messages with role + content display
- `reasoning_content` / `thinking` blocks shown before content
- Tool definitions (`tools`), tool calls (`tool_calls`), tool results (`tool_results`)
- SSE streaming content extraction (OpenAI + Anthropic delta events)
- Error detection for failed tool calls (JSON field check + word-boundary patterns)
- HTML content detection with DOM tree indentation
- Markdown rendering via markdown-it
