package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/llmscout/llmscout/internal/log"
	"github.com/llmscout/llmscout/internal/proxy"
	"github.com/llmscout/llmscout/internal/route"
	"github.com/llmscout/llmscout/internal/storage"
)

type App struct {
	ctx      context.Context
	routeMgr *route.Manager
	proxyEng *proxy.Engine
	logSvc   *log.Service
	settings *storage.SettingsRepo
	dbPath   string
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
	if a.logSvc != nil {
		a.logSvc.Close()
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
func (a *App) StartProxy() error        { return a.proxyEng.Start() }
func (a *App) StopProxy() error         { return a.proxyEng.Stop() }
func (a *App) ProxyStatus() proxy.Status { return a.proxyEng.Status() }

// Log methods
func (a *App) QueryLogs(filter log.Filter) (log.QueryResult, error) {
	return a.logSvc.Query(filter)
}

func (a *App) GetLog(id int64) (*log.Entry, error) { return a.logSvc.Get(id) }
func (a *App) ClearLogs() error                    { return a.logSvc.Clear() }

func (a *App) GetLogRouteNames() ([]string, error) { return a.logSvc.GetRouteNames() }

// Settings methods
func (a *App) GetSetting(key, defaultVal string) string {
	return a.settings.Get(key, defaultVal)
}

func (a *App) SetSetting(key, value string) error {
	return a.settings.Set(key, value)
}
