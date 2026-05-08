package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	stdlog "log"

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
	port    int
	server  *http.Server
	running bool
	startAt time.Time
	mu      sync.RWMutex
	matcher Matcher
	logSvc  interface {
		Record(log.Entry)
	}
}

func NewEngine(port int, matcher Matcher, logSvc interface {
	Record(log.Entry)
}) *Engine {
	return &Engine{
		port:    port,
		matcher: matcher,
		logSvc:  logSvc,
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
			stdlog.Printf("proxy server error: %v", err)
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
