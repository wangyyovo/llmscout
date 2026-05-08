# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Wails v2** desktop application — Go backend + Vue 3 frontend bundled into a native desktop app. The project was scaffolded from `wails-template-naive` and has not yet been customized beyond the template (the window title and module name are still `"projectname"`).

- **Backend**: Go 1.22 (`go.mod` root)
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
wails build --clean                  # Clean build
wails build --clean --platform windows/amd64   # Windows AMD64
wails build --clean --platform darwin/universal # macOS universal
wails build --clean --platform darwin/arm64     # macOS Apple Silicon
```

The `scripts/` directory contains platform-specific wrapper scripts that run `wails build --clean` with the appropriate `--platform` flag.

### Installing Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Architecture

### Wails Binding Pattern

The Go backend exposes methods to the frontend via Wails' `Bind` mechanism. Any method on a struct listed in `Bind` in `main.go` is callable from JavaScript.

```
main.go:
  Bind: []interface{}{app}  →  makes App struct methods callable from frontend

app.go:
  App.Greet(name string) string  →  exposed to frontend

frontend:
  import {Greet} from '../../wailsjs/go/main/App'
  Greet(data.name).then(result => { ... })
```

The `frontend/wailsjs/` directory is **auto-generated** by Wails at build time — do not edit it manually.

### Adding a New Backend Method

1. Add a method to `App` (or a new struct) in `app.go`
2. If using a new struct, add it to the `Bind` slice in `main.go`
3. Wails regenerates the JS bridge in `frontend/wailsjs/` on next build
4. Import and call it from a Vue component

### Frontend Structure

- **`frontend/src/main.js`** — Vue app bootstrap, registers NaiveUI plugin
- **`frontend/src/App.vue`** — Root component with `<n-message-provider>`, grid layout with logos, includes `<HelloWorld/>`
- **`frontend/src/components/`** — Vue components using `<script setup>` composition API
- **`frontend/vite.config.js`** — Minimal Vite config with Vue plugin only
- **`frontend/index.html`** — HTML entry point

### App Lifecycle Hooks

Defined in `main.go` → `app.go`:

| Hook | When | Purpose |
|------|------|---------|
| `startup(ctx)` | App launches, before frontend loads | One-time setup, DB connections, etc. |
| `domReady(ctx)` | Frontend DOM is fully loaded | Post-render initialization |
| `beforeClose(ctx)` | User clicks close / `runtime.Quit` | Return `true` to prevent closing |
| `shutdown(ctx)` | App is terminating | Cleanup, close connections |

### No Existing Infrastructure

The project currently has no tests, no linting, no CI/CD, no Docker, no environment variable configuration, and no TypeScript. These would need to be added from scratch.

## Go Module

- Module path: `projectname` (defined in `go.mod`)
- Primary dependency: `github.com/wailsapp/wails/v2 v2.12.0`
- Go embed directives in `main.go` bundle `frontend/dist` and `build/appicon.png` into the binary
