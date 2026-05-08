# LLM Scout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build LLM Scout — a desktop LLM debugging proxy with route-based forwarding, request/response logging, and SQLite persistence.

**Architecture:** Go backend (RouteManager + ProxyEngine + LogService + SQLite) exposed via Wails v2 Bind to a Vue 3 + NaiveUI frontend with collapsible sidebar layout.

**Tech Stack:** Go 1.22, Wails v2.12.0, Vue 3, NaiveUI, Vite 5, modernc.org/sqlite, net/http/httputil

---

### Task 1: Project Initialization

**Files:**
- Modify: `go.mod`
- Modify: `wails.json`
- Modify: `main.go`
- Delete: `frontend/src/components/HelloWorld.vue`
- Delete: `frontend/src/assets/images/logo-universal.png`
- Delete: `frontend/src/assets/images/naive-logo.svg`
- Delete: `frontend/src/style.css`
- Delete: `wails-naive.png`
- Modify: `frontend/src/App.vue`
- Create: `.gitignore` (only if missing)

- [ ] **Step 1: Update Go module name**

Edit `go.mod`, change `module projectname` to `module github.com/llmscout/llmscout`.

- [ ] **Step 2: Update Wails config**

Edit `wails.json`:
```json
{
  "name": "llmscout",
  "outputfilename": "llmscout",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "http://localhost:5173",
  "author": {
    "name": "wangyyovo",
    "email": "626562203@qq.com"
  }
}
```

- [ ] **Step 3: Update main.go title**

In `main.go`, change `Title: "projectname"` to `Title: "LLM Scout"` and `About: Title: "projectname"` to `About: Title: "LLM Scout"`.

- [ ] **Step 4: Clean up template files**

Delete these files:
```bash
rm frontend/src/components/HelloWorld.vue
rm frontend/src/assets/images/logo-universal.png
rm frontend/src/assets/images/naive-logo.svg
rm frontend/src/style.css
rm wails-naive.png
```

- [ ] **Step 5: Simplify App.vue**

Replace `frontend/src/App.vue` with a minimal sidebar layout shell:

```vue
<script setup>
</script>

<template>
  <n-message-provider>
    <n-layout has-sider position="absolute" style="height: 100vh;">
      <n-layout-sider
        bordered
        :collapsed="collapsed"
        collapse-mode="width"
        :collapsed-width="56"
        :width="180"
        :native-scrollbar="false"
        style="background: #1e1e2e;"
      >
        <n-menu
          :collapsed="collapsed"
          :collapsed-width="56"
          :collapsed-icon-size="20"
          :options="menuOptions"
        />
        <template #collapse-extra>
          <div style="padding: 8px; text-align: center; border-top: 1px solid #313244;">
            <n-button quaternary size="small" @click="collapsed = !collapsed">
              {{ collapsed ? '»' : '« 收缩' }}
            </n-button>
          </div>
        </template>
      </n-layout-sider>
      <n-layout content-style="padding: 0; background: #181825;">
        <router-view />
      </n-layout>
    </n-layout>
  </n-message-provider>
</template>

<style>
html, body { margin: 0; padding: 0; height: 100%; }
body { background: #181825; color: #cdd6f4; }
</style>
```

- [ ] **Step 6: Create internal package directories**

```bash
mkdir -p internal/route internal/proxy internal/log internal/storage
```

- [ ] **Step 7: Git init and first commit**

```bash
git init
git add -A
git commit -m "feat: initial project scaffold from wails-naive template

Rename to LLM Scout, clean up template files, set up project structure."
```

---

### Task 2: SQLite Storage Layer

**Files:**
- Create: `internal/storage/db.go`
- Create: `internal/storage/settings_repo.go`

- [ ] **Step 1: Add sqlite dependency**

```bash
cd /path/to/project
go get modernc.org/sqlite
```

- [ ] **Step 2: Create database initialization**

`internal/storage/db.go`:
```go
package storage

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL CHECK(type IN ('prefix','exact')),
		path TEXT NOT NULL,
		target_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_name TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		protocol TEXT NOT NULL DEFAULT 'REST',
		status_code INTEGER,
		latency_ms INTEGER,
		req_headers TEXT,
		req_body TEXT,
		resp_headers TEXT,
		resp_body TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_route_name ON logs(route_name);
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`
	_, err := db.Exec(schema)
	return err
}
```

- [ ] **Step 3: Create settings repository**

`internal/storage/settings_repo.go`:
```go
package storage

import "database/sql"

type SettingsRepo struct{ db *sql.DB }

func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(key, defaultVal string) string {
	var val string
	err := r.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	return val
}

func (r *SettingsRepo) Set(key, value string) error {
	_, err := r.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?",
		key, value, value,
	)
	return err
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/storage/ go.mod go.sum
git commit -m "feat: add sqlite storage layer with migrations"
```

---

### Task 3: RouteManager (Model + Repository)

**Files:**
- Create: `internal/route/model.go`
- Create: `internal/storage/routes_repo.go`
- Create: `internal/route/manager.go`

- [ ] **Step 1: Create route model**

`internal/route/model.go`:
```go
package route

import "time"

type Rule struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "prefix" or "exact"
	Path      string    `json:"path"`
	TargetURL string    `json:"targetUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
```

- [ ] **Step 2: Create routes repository**

`internal/storage/routes_repo.go`:
```go
package storage

import (
	"database/sql"
	"time"
	"github.com/llmscout/llmscout/internal/route"
)

type RoutesRepo struct{ db *sql.DB }

func NewRoutesRepo(db *sql.DB) *RoutesRepo {
	return &RoutesRepo{db: db}
}

func (r *RoutesRepo) List() ([]route.Rule, error) {
	rows, err := r.db.Query("SELECT id, name, type, path, target_url, created_at, updated_at FROM routes ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []route.Rule
	for rows.Next() {
		var rule route.Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Type, &rule.Path, &rule.TargetURL, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *RoutesRepo) Add(rule route.Rule) (int64, error) {
	now := time.Now()
	res, err := r.db.Exec(
		"INSERT INTO routes (name, type, path, target_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		rule.Name, rule.Type, rule.Path, rule.TargetURL, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RoutesRepo) Update(id int64, rule route.Rule) error {
	_, err := r.db.Exec(
		"UPDATE routes SET name=?, type=?, path=?, target_url=?, updated_at=? WHERE id=?",
		rule.Name, rule.Type, rule.Path, rule.TargetURL, time.Now(), id,
	)
	return err
}

func (r *RoutesRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM routes WHERE id=?", id)
	return err
}

func (r *RoutesRepo) Get(id int64) (*route.Rule, error) {
	var rule route.Rule
	err := r.db.QueryRow("SELECT id, name, type, path, target_url, created_at, updated_at FROM routes WHERE id=?", id).
		Scan(&rule.ID, &rule.Name, &rule.Type, &rule.Path, &rule.TargetURL, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
```

- [ ] **Step 3: Create RouteManager**

`internal/route/manager.go`:
```go
package route

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	rules []Rule
	repo  interface {
		List() ([]Rule, error)
		Add(Rule) (int64, error)
		Update(int64, Rule) error
		Delete(int64) error
		Get(int64) (*Rule, error)
	}
}

func NewManager(repo interface {
	List() ([]Rule, error)
	Add(Rule) (int64, error)
	Update(int64, Rule) error
	Delete(int64) error
	Get(int64) (*Rule, error)
}) *Manager {
	return &Manager{repo: repo}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rules, err := m.repo.List()
	if err != nil {
		return err
	}
	m.rules = rules
	return nil
}

func (m *Manager) List() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Rule, len(m.rules))
	copy(result, m.rules)
	return result
}

func (m *Manager) Add(rule Rule) (int64, error) {
	id, err := m.repo.Add(rule)
	if err != nil {
		return 0, err
	}
	rule.ID = id
	m.mu.Lock()
	m.rules = append(m.rules, rule)
	m.mu.Unlock()
	return id, nil
}

func (m *Manager) Update(id int64, rule Rule) error {
	if err := m.repo.Update(id, rule); err != nil {
		return err
	}
	rule.ID = id
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rules {
		if r.ID == id {
			m.rules[i] = rule
			return nil
		}
	}
	return fmt.Errorf("route %d not found", id)
}

func (m *Manager) Delete(id int64) error {
	if err := m.repo.Delete(id); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rules {
		if r.ID == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return nil
}

// Match returns the target URL and whether a match was found.
// For prefix rules, strips the path prefix and appends remaining path to target URL.
// For exact rules, returns the full target URL as-is.
func (m *Manager) Match(requestPath string) (targetURL string, routeName string, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, rule := range m.rules {
		switch rule.Type {
		case "exact":
			if requestPath == rule.Path {
				return rule.TargetURL, rule.Name, true
			}
		case "prefix":
			if strings.HasPrefix(requestPath, rule.Path) {
				suffix := strings.TrimPrefix(requestPath, rule.Path)
				return strings.TrimRight(rule.TargetURL, "/") + "/" + strings.TrimLeft(suffix, "/"), rule.Name, true
			}
		}
	}
	return "", "", false
}

// IsValidURL checks that a target URL has a valid scheme and host.
func IsValidURL(target string) bool {
	u, err := url.Parse(target)
	return err == nil && u.Scheme != "" && u.Host != ""
}
```

- [ ] **Step 4: Create go.mod entries**

Run: `cd /path/to/project && go mod tidy`

- [ ] **Step 5: Commit**

```bash
git add internal/route/ internal/storage/routes_repo.go
git commit -m "feat: add route manager with CRUD and path matching"
```

---

### Task 4: LogService (Model + Repository + Service)

**Files:**
- Create: `internal/log/model.go`
- Create: `internal/storage/logs_repo.go`
- Create: `internal/log/service.go`

- [ ] **Step 1: Create log model**

`internal/log/model.go`:
```go
package log

import "time"

type Entry struct {
	ID         int64     `json:"id"`
	RouteName  string    `json:"routeName"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Protocol   string    `json:"protocol"` // "REST" | "SSE"
	StatusCode int       `json:"statusCode"`
	LatencyMs  int64     `json:"latencyMs"`
	ReqHeaders string    `json:"reqHeaders"`
	ReqBody    string    `json:"reqBody"`
	RespHeaders string   `json:"respHeaders"`
	RespBody   string    `json:"respBody"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Filter struct {
	Keyword    string `json:"keyword"`
	RouteName  string `json:"routeName"`
	StatusCode int    `json:"statusCode"` // 0 means all
	Protocol   string `json:"protocol"`   // empty means all
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

type QueryResult struct {
	List  []Entry `json:"list"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
}
```

- [ ] **Step 2: Create logs repository**

`internal/storage/logs_repo.go`:
```go
package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"log"
	"github.com/llmscout/llmscout/internal/log"
)

type LogsRepo struct{ db *sql.DB }

func NewLogsRepo(db *sql.DB) *LogsRepo {
	return &LogsRepo{db: db}
}

func (r *LogsRepo) Insert(entry log.Entry) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO logs (route_name, method, path, protocol, status_code, latency_ms,
		 req_headers, req_body, resp_headers, resp_body, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.RouteName, entry.Method, entry.Path, entry.Protocol,
		entry.StatusCode, entry.LatencyMs, entry.ReqHeaders, entry.ReqBody,
		entry.RespHeaders, entry.RespBody, entry.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *LogsRepo) Query(filter log.Filter) (log.QueryResult, error) {
	var where []string
	var args []interface{}

	if filter.RouteName != "" {
		where = append(where, "route_name = ?")
		args = append(args, filter.RouteName)
	}
	if filter.StatusCode > 0 {
		where = append(where, "status_code = ?")
		args = append(args, filter.StatusCode)
	}
	if filter.Protocol != "" {
		where = append(where, "protocol = ?")
		args = append(args, filter.Protocol)
	}
	if filter.StartTime != "" {
		where = append(where, "created_at >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		where = append(where, "created_at <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.Keyword != "" {
		where = append(where, "(req_body LIKE ? OR resp_body LIKE ?)")
		kw := "%" + filter.Keyword + "%"
		args = append(args, kw, kw)
	}

	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM logs" + clause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return log.QueryResult{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	dataQuery := fmt.Sprintf("SELECT id, route_name, method, path, protocol, status_code, latency_ms, req_headers, req_body, resp_headers, resp_body, created_at FROM logs%s ORDER BY id DESC LIMIT ? OFFSET ?", clause)
	dataArgs := append(args, filter.PageSize, offset)
	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return log.QueryResult{}, err
	}
	defer rows.Close()

	var entries []log.Entry
	for rows.Next() {
		var e log.Entry
		if err := rows.Scan(&e.ID, &e.RouteName, &e.Method, &e.Path, &e.Protocol,
			&e.StatusCode, &e.LatencyMs, &e.ReqHeaders, &e.ReqBody,
			&e.RespHeaders, &e.RespBody, &e.CreatedAt); err != nil {
			log.Printf("scan log row: %v", err)
			continue
		}
		entries = append(entries, e)
	}
	return log.QueryResult{List: entries, Total: total, Page: filter.Page}, nil
}

func (r *LogsRepo) Get(id int64) (*log.Entry, error) {
	var e log.Entry
	err := r.db.QueryRow(
		"SELECT id, route_name, method, path, protocol, status_code, latency_ms, req_headers, req_body, resp_headers, resp_body, created_at FROM logs WHERE id=?", id,
	).Scan(&e.ID, &e.RouteName, &e.Method, &e.Path, &e.Protocol,
		&e.StatusCode, &e.LatencyMs, &e.ReqHeaders, &e.ReqBody,
		&e.RespHeaders, &e.RespBody, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *LogsRepo) DeleteAll() error {
	_, err := r.db.Exec("DELETE FROM logs")
	return err
}

func (r *LogsRepo) GetRouteNames() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT route_name FROM logs ORDER BY route_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, nil
}
```

- [ ] **Step 3: Create LogService**

`internal/log/service.go`:
```go
package log

import "time"

type entryRepo interface {
	Insert(Entry) (int64, error)
	Query(Filter) (QueryResult, error)
	Get(int64) (*Entry, error)
	DeleteAll() error
	GetRouteNames() ([]string, error)
}

type Service struct {
	repo    entryRepo
	entries chan Entry
}

func NewService(repo entryRepo) *Service {
	s := &Service{
		repo:    repo,
		entries: make(chan Entry, 100),
	}
	go s.worker()
	return s
}

func (s *Service) worker() {
	for entry := range s.entries {
		entry.CreatedAt = time.Now()
		_, err := s.repo.Insert(entry)
		if err != nil {
			// Log and drop — never block the proxy
		}
	}
}

// Record enqueues an entry for async storage. Drops if buffer is full.
func (s *Service) Record(entry Entry) {
	select {
	case s.entries <- entry:
	default:
	}
}

func (s *Service) Query(filter Filter) (QueryResult, error) {
	return s.repo.Query(filter)
}

func (s *Service) Get(id int64) (*Entry, error) {
	return s.repo.Get(id)
}

func (s *Service) Clear() error {
	return s.repo.DeleteAll()
}

func (s *Service) GetRouteNames() ([]string, error) {
	return s.repo.GetRouteNames()
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/log/ internal/storage/logs_repo.go
git commit -m "feat: add log service with async writes and query with filtering"
```

---

### Task 5: ProxyEngine

**Files:**
- Create: `internal/proxy/engine.go`

- [ ] **Step 1: Create the proxy engine**

`internal/proxy/engine.go`:
```go
package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"
	"github.com/llmscout/llmscout/internal/log"
)

type Status struct {
	Running   bool      `json:"running"`
	Port      int       `json:"port"`
	StartTime time.Time `json:"startTime"`
	Uptime    string    `json:"uptime"`
}

type Matcher interface {
	Match(path string) (targetURL string, routeName string, ok bool)
}

type Engine struct {
	port     int
	server   *http.Server
	running  bool
	startAt  time.Time
	mu       sync.RWMutex
	matcher  Matcher
	logSvc   interface {
		Record(log.Entry)
	}
	// For SSE streaming
	sseBuf sync.Pool
}

func NewEngine(port int, matcher Matcher, logSvc interface {
	Record(log.Entry)
}) *Engine {
	return &Engine{
		port:    port,
		matcher: matcher,
		logSvc:  logSvc,
		sseBuf: sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
}

func (e *Engine) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("proxy already running")
	}
	e.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/", e.handleRequest)

	e.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", e.port),
		Handler: mux,
	}

	e.mu.Lock()
	e.running = true
	e.startAt = time.Now()
	e.mu.Unlock()

	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("proxy server error: %v", err)
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
		}
	}()

	return nil
}

func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return nil
	}
	if e.server != nil {
		e.server.Close()
	}
	e.running = false
	return nil
}

func (e *Engine) Status() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	uptime := ""
	if e.running {
		uptime = time.Since(e.startAt).Round(time.Second).String()
	}
	return Status{
		Running:   e.running,
		Port:      e.port,
		StartTime: e.startAt,
		Uptime:    uptime,
	}
}

func (e *Engine) SetPort(port int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.port = port
}

func (e *Engine) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	targetURL, routeName, ok := e.matcher.Match(r.URL.Path)
	if !ok {
		http.Error(w, "No matching route for path: "+r.URL.Path, http.StatusBadGateway)
		e.logSvc.Record(log.Entry{
			RouteName:  "unknown",
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: 502,
			LatencyMs:  time.Since(start).Milliseconds(),
			Protocol:   "REST",
		})
		return
	}

	// Read and store request body
	reqBody, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(reqBody))

	// Capture request headers as JSON
	reqHeaders := headersToJSON(r.Header)

	// Build outgoing request
	outReq, _ := http.NewRequest(r.Method, targetURL, bytes.NewReader(reqBody))
	outReq.Header = r.Header.Clone()
	outReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, "Upstream error: "+err.Error(), http.StatusBadGateway)
		e.logSvc.Record(log.Entry{
			RouteName:  routeName,
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: 502,
			LatencyMs:  time.Since(start).Milliseconds(),
			ReqHeaders: reqHeaders,
			ReqBody:    string(reqBody),
			Protocol:   "REST",
		})
		return
	}
	defer resp.Body.Close()

	respHeaders := headersToJSON(resp.Header)
	contentType := resp.Header.Get("Content-Type")

	// Copy response headers to client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	if strings.Contains(contentType, "text/event-stream") {
		// SSE: stream to client while buffering
		var sseBuf bytes.Buffer
		written, _ := io.Copy(io.MultiWriter(w, &sseBuf), resp.Body)

		e.logSvc.Record(log.Entry{
			RouteName:   routeName,
			Method:      r.Method,
			Path:        r.URL.Path,
			StatusCode:  resp.StatusCode,
			LatencyMs:   time.Since(start).Milliseconds(),
			ReqHeaders:  reqHeaders,
			ReqBody:     string(reqBody),
			RespHeaders: respHeaders,
			RespBody:    sseBuf.String(),
			Protocol:    "SSE",
		})
		_ = written
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		w.Write(respBody)

		e.logSvc.Record(log.Entry{
			RouteName:   routeName,
			Method:      r.Method,
			Path:        r.URL.Path,
			StatusCode:  resp.StatusCode,
			LatencyMs:   time.Since(start).Milliseconds(),
			ReqHeaders:  reqHeaders,
			ReqBody:     string(reqBody),
			RespHeaders: respHeaders,
			RespBody:    string(respBody),
			Protocol:    "REST",
		})
	}
}

func headersToJSON(h http.Header) string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for k, v := range h {
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteString(fmt.Sprintf("%q:%q", k, strings.Join(v, ", ")))
	}
	b.WriteByte('}')
	return b.String()
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/proxy/
go mod tidy
git add go.mod go.sum
git commit -m "feat: add proxy engine with SSE support and async logging"
```

---

### Task 6: App Integration (Wails Bind Layer)

**Files:**
- Modify: `app.go`
- Modify: `main.go`

- [ ] **Step 1: Rewrite app.go**

`app.go`:
```go
package main

import (
	"context"
	"strconv"
	"os"
	"path/filepath"

	"github.com/llmscout/llmscout/internal/log"
	"github.com/llmscout/llmscout/internal/proxy"
	"github.com/llmscout/llmscout/internal/route"
	"github.com/llmscout/llmscout/internal/storage"
)

type App struct {
	ctx       context.Context
	routeMgr  *route.Manager
	proxyEng  *proxy.Engine
	logSvc    *log.Service
	settings  *storage.SettingsRepo
	dbPath    string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Determine DB path
	home, _ := os.UserHomeDir()
	a.dbPath = filepath.Join(home, ".llmscout", "data.db")
	os.MkdirAll(filepath.Dir(a.dbPath), 0755)

	// Init DB
	db, err := storage.InitDB(a.dbPath)
	if err != nil {
		panic("failed to init db: " + err.Error())
	}

	// Init repos and services
	routesRepo := storage.NewRoutesRepo(db)
	a.settings = storage.NewSettingsRepo(db)
	logsRepo := storage.NewLogsRepo(db)
	a.logSvc = log.NewService(logsRepo)

	a.routeMgr = route.NewManager(routesRepo)
	if err := a.routeMgr.Load(); err != nil {
		panic("failed to load routes: " + err.Error())
	}

	portStr := a.settings.Get("port", "8899")
	port := 8899
	if p, err := strconv.Atoi(portStr); err == nil {
		port = p
	}
	a.proxyEng = proxy.NewEngine(port, a.routeMgr, a.logSvc)
}

func (a *App) shutdown(ctx context.Context) {
	if a.proxyEng != nil {
		a.proxyEng.Stop()
	}
}

func (a *App) domReady(ctx context.Context) {}
func (a *App) beforeClose(ctx context.Context) (prevent bool) { return false }

// Route methods
func (a *App) ListRoutes() []route.Rule { return a.routeMgr.List() }

func (a *App) AddRoute(name, ruleType, path, targetURL string) (int64, error) {
	return a.routeMgr.Add(route.Rule{
		Name:      name,
		Type:      ruleType,
		Path:      path,
		TargetURL: targetURL,
	})
}

func (a *App) UpdateRoute(id int64, name, ruleType, path, targetURL string) error {
	return a.routeMgr.Update(id, route.Rule{
		Name:      name,
		Type:      ruleType,
		Path:      path,
		TargetURL: targetURL,
	})
}

func (a *App) DeleteRoute(id int64) error { return a.routeMgr.Delete(id) }

// Proxy methods
func (a *App) StartProxy() error  { return a.proxyEng.Start() }
func (a *App) StopProxy() error   { return a.proxyEng.Stop() }
func (a *App) ProxyStatus() proxy.Status { return a.proxyEng.Status() }

// Log methods
func (a *App) QueryLogs(filter log.Filter) (log.QueryResult, error) {
	return a.logSvc.Query(filter)
}

func (a *App) GetLog(id int64) (*log.Entry, error) { return a.logSvc.Get(id) }
func (a *App) ClearLogs() error                    { return a.logSvc.Clear() }

func (a *App) GetLogRouteNames() ([]string, error)  { return a.logSvc.GetRouteNames() }

// Settings methods
func (a *App) GetSetting(key, defaultVal string) string {
	return a.settings.Get(key, defaultVal)
}

func (a *App) SetSetting(key, value string) error {
	return a.settings.Set(key, value)
}
```

- [ ] **Step 2: Build and verify compilation**

```bash
cd /path/to/project
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add app.go main.go
git commit -m "feat: integrate all services via Wails Bind layer"
```

---

### Task 7: Frontend — Sidebar Layout and Routing

**Files:**
- Modify: `frontend/src/App.vue`
- Create: `frontend/src/views/ProxyPanel.vue`
- Create: `frontend/src/views/RoutePanel.vue`
- Create: `frontend/src/views/LogViewer.vue`
- Create: `frontend/src/views/SettingsPanel.vue`
- Modify: `frontend/src/main.js`

- [ ] **Step 1: Update main.js to set up NaiveUI**

`frontend/src/main.js`:
```js
import { createApp } from 'vue'
import naive from 'naive-ui'
import App from './App.vue'

const app = createApp(App)
app.use(naive)
app.mount('#app')
```

- [ ] **Step 2: Create App.vue with sidebar layout and view switching**

`frontend/src/App.vue`:
```vue
<script setup>
import { ref, shallowRef } from 'vue'
import { NMessageProvider, NLayout, NLayoutSider, NMenu, NButton } from 'naive-ui'
import ProxyPanel from './views/ProxyPanel.vue'
import RoutePanel from './views/RoutePanel.vue'
import LogViewer from './views/LogViewer.vue'
import SettingsPanel from './views/SettingsPanel.vue'

const collapsed = ref(false)
const activeTab = ref('proxy')

const menuOptions = [
  { label: () => '代理', key: 'proxy', icon: () => '📡' },
  { label: () => '路由', key: 'routes', icon: () => '🔀' },
  { label: () => '日志', key: 'logs', icon: () => '📋' },
  { label: () => '设置', key: 'settings', icon: () => '⚙' },
]

const currentView = shallowRef(ProxyPanel)

function handleUpdate(key) {
  activeTab.value = key
  const views = { proxy: ProxyPanel, routes: RoutePanel, logs: LogViewer, settings: SettingsPanel }
  currentView.value = views[key] || ProxyPanel
}
</script>

<template>
  <n-message-provider>
    <n-layout has-sider position="absolute" style="height: 100vh;">
      <n-layout-sider
        bordered
        :collapsed="collapsed"
        collapse-mode="width"
        :collapsed-width="56"
        :width="180"
        :native-scrollbar="false"
        style="background: #1e1e2e;"
      >
        <n-menu
          :collapsed="collapsed"
          :collapsed-width="56"
          :collapsed-icon-size="20"
          :options="menuOptions"
          :value="activeTab"
          @update:value="handleUpdate"
        />
        <template #collapse-extra>
          <div style="padding: 8px; text-align: center; border-top: 1px solid #313244;">
            <n-button quaternary size="small" @click="collapsed = !collapsed" style="color: #6c7086;">
              {{ collapsed ? '»' : '« 收缩' }}
            </n-button>
          </div>
        </template>
      </n-layout-sider>
      <n-layout content-style="padding: 20px 24px; background: #181825; color: #cdd6f4;">
        <component :is="currentView" />
      </n-layout>
    </n-layout>
  </n-message-provider>
</template>

<style>
html, body { margin: 0; padding: 0; height: 100%; }
body { background: #181825; }
</style>
```

- [ ] **Step 3: Build frontend to verify**

```bash
cd frontend && npm install && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/
git commit -m "feat: add sidebar layout with view switching"
```

---

### Task 8: Frontend — ProxyPanel and SettingsPanel

**Files:**
- Create: `frontend/src/views/ProxyPanel.vue`
- Create: `frontend/src/views/SettingsPanel.vue`

- [ ] **Step 1: Create ProxyPanel**

`frontend/src/views/ProxyPanel.vue`:
```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NInput, NButton, NCard, NStatistic, NSpace, NTag } from 'naive-ui'
import { StartProxy, StopProxy, ProxyStatus, ListRoutes, QueryLogs, GetSetting } from '../../wailsjs/go/main/App'

const port = ref(8899)
const running = ref(false)
const uptime = ref('')
const routeCount = ref(0)
const todayCount = ref(0)

let statusTimer = null

async function refreshStatus() {
  const s = await ProxyStatus()
  running.value = s.running
  uptime.value = s.uptime || ''
  port.value = s.port
  const routes = await ListRoutes()
  routeCount.value = routes.length
  const result = await QueryLogs({ keyword: '', routeName: '', statusCode: 0, protocol: '', startTime: '', endTime: '', page: 1, pageSize: 1 })
  todayCount.value = Number(result.total) || 0
}

async function toggleProxy() {
  if (running.value) {
    await StopProxy()
  } else {
    await StartProxy()
  }
  refreshStatus()
}

onMounted(async () => {
  const savedPort = await GetSetting('port', '8899')
  port.value = parseInt(savedPort) || 8899
  refreshStatus()
  statusTimer = setInterval(refreshStatus, 3000)
})

onUnmounted(() => {
  if (statusTimer) clearInterval(statusTimer)
})
</script>

<template>
  <div>
    <h2 style="color: #cdd6f4; margin-bottom: 20px;">📡 代理控制</h2>
    <n-card style="background: #1e1e2e; border: none; max-width: 500px;">
      <n-space vertical size="large">
        <div style="display: flex; align-items: center; gap: 12px;">
          <span style="color: #a6adc8;">端口</span>
          <n-input
            v-model:value="port"
            :disabled="running"
            style="width: 120px;"
            type="number"
          />
          <span style="color: #6c7086; font-size: 13px;">localhost:{{ port }}</span>
        </div>
        <div style="display: flex; align-items: center; gap: 12px;">
          <span style="color: #a6adc8;">状态</span>
          <n-tag :type="running ? 'success' : 'default'" size="small">
            {{ running ? '● 运行中' : '已停止' }}
          </n-tag>
          <n-button
            :type="running ? 'error' : 'primary'"
            @click="toggleProxy"
          >
            {{ running ? '停止代理' : '启动代理' }}
          </n-button>
        </div>
      </n-space>
    </n-card>

    <div style="display: flex; gap: 20px; margin-top: 20px;">
      <n-card style="background: #1e1e2e; border: none; flex: 1;" size="small">
        <n-statistic label="已配置路由" :value="routeCount" />
      </n-card>
      <n-card style="background: #1e1e2e; border: none; flex: 1;" size="small">
        <n-statistic label="今日请求" :value="todayCount" />
      </n-card>
      <n-card style="background: #1e1e2e; border: none; flex: 1;" size="small">
        <n-statistic label="运行时长" :value="uptime || '-'" />
      </n-card>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Create SettingsPanel**

`frontend/src/views/SettingsPanel.vue`:
```vue
<script setup>
import { ref, onMounted } from 'vue'
import { NInput, NButton, NCard, NSpace, useMessage } from 'naive-ui'
import { GetSetting, SetSetting, ClearLogs } from '../../wailsjs/go/main/App'

const message = useMessage()
const port = ref('8899')
const dbPath = ref('')

onMounted(async () => {
  port.value = await GetSetting('port', '8899')
  dbPath.value = await GetSetting('dbPath', '~/.llmscout/data.db')
})

async function savePort() {
  await SetSetting('port', port.value)
  message.success('端口设置已保存，下次启动代理时生效')
}

async function handleClear() {
  await ClearLogs()
  message.success('日志已清空')
}
</script>

<template>
  <div>
    <h2 style="color: #cdd6f4; margin-bottom: 20px;">⚙ 设置</h2>
    <n-card style="background: #1e1e2e; border: none; max-width: 500px;">
      <n-space vertical size="large">
        <div>
          <div style="color: #a6adc8; margin-bottom: 6px;">代理端口</div>
          <div style="display: flex; gap: 10px;">
            <n-input v-model:value="port" type="number" style="width: 120px;" />
            <n-button type="primary" @click="savePort">保存</n-button>
          </div>
        </div>
        <div>
          <div style="color: #a6adc8; margin-bottom: 6px;">数据库路径</div>
          <div style="color: #6c7086; font-size: 13px;">{{ dbPath }}</div>
        </div>
        <n-button type="error" @click="handleClear">清空所有日志</n-button>
      </n-space>
    </n-card>
  </div>
</template>
```

- [ ] **Step 3: Build frontend to verify**

```bash
cd frontend && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/ProxyPanel.vue frontend/src/views/SettingsPanel.vue
git commit -m "feat: add proxy control and settings panels"
```

---

### Task 9: Frontend — RoutePanel

**Files:**
- Create: `frontend/src/views/RoutePanel.vue`

- [ ] **Step 1: Create RoutePanel**

`frontend/src/views/RoutePanel.vue`:
```vue
<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NSpace, NTag, NModal, NInput, NSelect, useMessage } from 'naive-ui'
import { ListRoutes, AddRoute, UpdateRoute, DeleteRoute } from '../../wailsjs/go/main/App'

const message = useMessage()
const rules = ref([])
const showModal = ref(false)
const editingId = ref(null)

const form = ref({ name: '', type: 'prefix', path: '', targetUrl: '' })
const typeOptions = [
  { label: 'prefix — 路径前缀剥离', value: 'prefix' },
  { label: 'exact — 精确路径映射', value: 'exact' },
]

async function load() {
  rules.value = await ListRoutes()
}

function openAdd() {
  editingId.value = null
  form.value = { name: '', type: 'prefix', path: '', targetUrl: '' }
  showModal.value = true
}

function openEdit(rule) {
  editingId.value = rule.id
  form.value = { name: rule.name, type: rule.type, path: rule.path, targetUrl: rule.targetUrl }
  showModal.value = true
}

async function save() {
  if (!form.value.name || !form.value.path || !form.value.targetUrl) {
    message.error('请填完所有字段')
    return
  }
  if (editingId.value) {
    await UpdateRoute(editingId.value, form.value.name, form.value.type, form.value.path, form.value.targetUrl)
    message.success('路由已更新')
  } else {
    await AddRoute(form.value.name, form.value.type, form.value.path, form.value.targetUrl)
    message.success('路由已添加')
  }
  showModal.value = false
  load()
}

async function remove(id) {
  await DeleteRoute(id)
  message.success('路由已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="color: #cdd6f4;">🔀 路由规则</h2>
      <n-button type="primary" @click="openAdd">+ 添加路由</n-button>
    </div>

    <n-card v-for="rule in rules" :key="rule.id"
      style="background: #1e1e2e; border: none; margin-bottom: 8px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <n-tag :type="rule.type === 'prefix' ? 'info' : 'warning'" size="small">
          {{ rule.type }}
        </n-tag>
        <code style="color: #cdd6f4;">{{ rule.path }}</code>
        <span style="color: #6c7086;">→</span>
        <code style="color: #a6e3a1;">{{ rule.targetUrl }}</code>
        <span style="color: #6c7086; font-size: 12px; margin-left: 4px;">({{ rule.name }})</span>
        <span style="margin-left: auto;">
          <n-button quaternary size="small" @click="openEdit(rule)" style="color: #6c7086;">编辑</n-button>
          <n-button quaternary size="small" @click="remove(rule.id)" style="color: #f38ba8;">删除</n-button>
        </span>
      </div>
    </n-card>

    <div v-if="rules.length === 0" style="color: #6c7086; text-align: center; padding: 40px;">
      暂无路由规则，点击上方按钮添加
    </div>

    <n-modal v-model:show="showModal" preset="card" title="路由规则" style="max-width: 500px;">
      <n-space vertical size="large">
        <n-input v-model:value="form.name" placeholder="名称（如 openai）" />
        <n-select v-model:value="form.type" :options="typeOptions" />
        <n-input v-model:value="form.path" placeholder="代理路径（如 /openai）" />
        <n-input v-model:value="form.targetUrl" placeholder="目标域名（如 api.openai.com）" />
        <n-button type="primary" @click="save">保存</n-button>
      </n-space>
    </n-modal>
  </div>
</template>
```

- [ ] **Step 2: Build frontend to verify**

```bash
cd frontend && npm run build
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/RoutePanel.vue
git commit -m "feat: add route management panel with CRUD"
```

---

### Task 10: Frontend — LogViewer with JsonViewer

**Files:**
- Create: `frontend/src/views/LogViewer.vue`
- Create: `frontend/src/components/JsonViewer.vue`

- [ ] **Step 1: Create JsonViewer component**

`frontend/src/components/JsonViewer.vue`:
```vue
<script setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'

const props = defineProps({ data: { type: String, default: '' } })

const formatted = computed(() => {
  if (!props.data) return ''
  try {
    return JSON.stringify(JSON.parse(props.data), null, 2)
  } catch {
    return props.data
  }
})

function copy() {
  navigator.clipboard.writeText(formatted.value)
}
</script>

<template>
  <div style="position: relative;">
    <div style="position: absolute; top: 8px; right: 8px;">
      <n-button quaternary size="tiny" @click="copy" style="color: #6c7086;">📋</n-button>
    </div>
    <pre style="background: #11111b; border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: #cdd6f4;"><code>{{ formatted }}</code></pre>
  </div>
</template>
```

- [ ] **Step 2: Create LogViewer**

`frontend/src/views/LogViewer.vue`:
```vue
<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { NInput, NSelect, NButton, NTag, NTable, NPagination, NSwitch, NModal, NTabs, NTabPane } from 'naive-ui'
import { QueryLogs, GetLog, GetLogRouteNames } from '../../wailsjs/go/main/App'
import JsonViewer from '../components/JsonViewer.vue'

const logs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const autoRefresh = ref(false)
const refreshInterval = ref(3)
const detailLog = ref(null)
const showDetail = ref(false)

const keyword = ref('')
const routeName = ref('')
const statusCode = ref(null)
const protocol = ref('')
const routeOptions = ref([])

let timer = null

const protocolOptions = [
  { label: '全部协议', value: '' },
  { label: 'REST', value: 'REST' },
  { label: 'SSE', value: 'SSE' },
]

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '2xx', value: 2 },
  { label: '3xx', value: 3 },
  { label: '4xx', value: 4 },
  { label: '5xx', value: 5 },
]

async function load() {
  const filter = {
    keyword: keyword.value,
    routeName: routeName.value,
    statusCode: statusCode.value ? statusCode.value * 100 : 0,
    protocol: protocol.value,
    startTime: '',
    endTime: '',
    page: page.value,
    pageSize: pageSize.value,
  }
  try {
    const result = await QueryLogs(filter)
    logs.value = result.list || []
    total.value = result.total || 0
  } catch (e) {
    // silent
  }
}

async function loadRouteNames() {
  try {
    const names = await GetLogRouteNames()
    routeOptions.value = [{ label: '全部服务商', value: '' }, ...names.map(n => ({ label: n, value: n }))]
  } catch {
    routeOptions.value = [{ label: '全部服务商', value: '' }]
  }
}

function search() {
  page.value = 1
  load()
}

function toggleAuto(v) {
  autoRefresh.value = v
  if (v) {
    timer = setInterval(load, refreshInterval.value * 1000)
  } else if (timer) {
    clearInterval(timer)
    timer = null
  }
}

async function openDetail(id) {
  try {
    detailLog.value = await GetLog(id)
    showDetail.value = true
  } catch { /* ignore */ }
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const pad = n => String(n).padStart(2, '0')
  return `${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function statusTagType(code) {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'info'
  if (code >= 400 && code < 500) return 'warning'
  if (code >= 500) return 'error'
  return 'default'
}

onMounted(() => {
  load()
  loadRouteNames()
})
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <div>
    <h2 style="color: #cdd6f4; margin-bottom: 16px;">📋 请求日志</h2>

    <!-- Filters -->
    <div style="display: flex; gap: 10px; margin-bottom: 14px; align-items: center; flex-wrap: wrap;">
      <n-input v-model:value="keyword" placeholder="🔍 搜索关键词..." clearable style="width: 180px;" @keyup.enter="search" />
      <n-select v-model:value="routeName" :options="routeOptions" style="width: 130px;" @update:value="search" />
      <n-select v-model:value="statusCode" :options="statusOptions" style="width: 110px;" @update:value="search" />
      <n-select v-model:value="protocol" :options="protocolOptions" style="width: 110px;" @update:value="search" />
      <n-button type="primary" size="small" @click="search">搜索</n-button>
      <span style="margin-left: auto; display: flex; align-items: center; gap: 6px; color: #a6adc8; font-size: 13px;">
        <span>🔄</span>
        <n-switch v-model:value="autoRefresh" @update:value="toggleAuto" />
        <n-select v-model:value="refreshInterval" :options="[{label:'3 秒',value:3},{label:'5 秒',value:5},{label:'10 秒',value:10}]" style="width: 80px;" />
      </span>
    </div>

    <!-- Table -->
    <div style="background: #1e1e2e; border-radius: 8px; overflow: hidden;">
      <n-table :single-line="false" style="background: transparent;">
        <thead>
          <tr style="background: #313244;">
            <th style="color: #a6adc8; width: 55px;">协议</th>
            <th style="color: #a6adc8; width: 55px;">方法</th>
            <th style="color: #a6adc8; width: 65px;">状态</th>
            <th style="color: #a6adc8; width: 90px;">服务商</th>
            <th style="color: #a6adc8;">路径</th>
            <th style="color: #a6adc8; width: 65px;">耗时</th>
            <th style="color: #a6adc8; width: 140px;">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" @click="openDetail(log.id)" style="cursor: pointer;">
            <td><n-tag :type="log.protocol === 'SSE' ? 'warning' : 'info'" size="tiny">{{ log.protocol }}</n-tag></td>
            <td style="color: #89b4fa;">{{ log.method }}</td>
            <td><n-tag :type="statusTagType(log.statusCode)" size="tiny">{{ log.statusCode }}</n-tag></td>
            <td style="color: #cdd6f4;">{{ log.routeName }}</td>
            <td><code style="color: #a6e3a1; font-size: 12px;">{{ log.path }}</code></td>
            <td style="color: #cdd6f4;">{{ log.latencyMs }}ms</td>
            <td style="color: #a6adc8; font-size: 12px;">{{ formatTime(log.createdAt) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="7" style="text-align: center; color: #6c7086; padding: 40px;">暂无日志记录</td>
          </tr>
        </tbody>
      </n-table>
    </div>

    <!-- Pagination -->
    <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 12px;">
      <div style="color: #a6adc8; font-size: 12px;">
        每页 <n-select v-model:value="pageSize" :options="[{label:'20',value:20},{label:'50',value:50},{label:'100',value:100}]" style="width: 70px; display: inline-block;" @update:value="load" /> 条
      </div>
      <n-pagination
        v-model:page="page"
        :page-count="Math.ceil(total / pageSize)"
        @update:page="load"
        simple
      />
    </div>

    <!-- Detail Modal -->
    <n-modal v-model:show="showDetail" preset="card" title="请求详情" style="max-width: 800px;" :segmented="{ content: true }">
      <template v-if="detailLog">
        <n-tabs type="line">
          <n-tab-pane name="reqBody" tab="请求体">
            <json-viewer :data="detailLog.reqBody" />
          </n-tab-pane>
          <n-tab-pane name="reqHeaders" tab="请求头">
            <json-viewer :data="detailLog.reqHeaders" />
          </n-tab-pane>
          <n-tab-pane name="respBody" tab="响应体">
            <json-viewer :data="detailLog.respBody" />
          </n-tab-pane>
          <n-tab-pane name="respHeaders" tab="响应头">
            <json-viewer :data="detailLog.respHeaders" />
          </n-tab-pane>
        </n-tabs>
      </template>
    </n-modal>
  </div>
</template>
```

- [ ] **Step 2: Build frontend to verify**

```bash
cd frontend && npm run build
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/LogViewer.vue frontend/src/components/JsonViewer.vue
git commit -m "feat: add log viewer with filters, pagination, auto-refresh and JSON viewer"
```

---

### Task 11: Verify Full Application

**Files:** (none — verification only)

- [ ] **Step 1: Build Go backend**

```bash
cd /path/to/project
go build ./...
```

Output expected: no errors.

- [ ] **Step 2: Run Wails dev mode**

```bash
wails dev
```

Expected: application window opens with sidebar, "代理" tab active.

- [ ] **Step 3: Test proxy functionality (manual)**

1. Add a route: type=`prefix`, path=`/openai`, target=`https://api.openai.com`
2. Click "启动代理" on the proxy panel
3. Send a test request:
```bash
curl http://localhost:8899/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-test" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}'
```
4. Switch to "日志" tab, verify the request appears in the table
5. Click the log row, verify the detail modal shows formatted JSON

- [ ] **Step 4: Commit final adjustments**

```bash
git add -A
git commit -m "chore: final adjustments after integration verification"
```
