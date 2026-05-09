package proxy

import (
	"bytes"
	"context"
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
	Running   bool   `json:"running"`
	Port      int    `json:"port"`
	StartTime int64  `json:"startTime"` // UnixMilli
	Uptime    string `json:"uptime"`
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
	client *http.Client
}

func NewEngine(port int, matcher Matcher, logSvc interface {
	Record(log.Entry)
}) *Engine {
	return &Engine{
		port:    port,
		matcher: matcher,
		logSvc:  logSvc,
		client:  &http.Client{Timeout: 180 * time.Second},
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := e.server.Shutdown(ctx); err != nil {
			stdlog.Printf("proxy shutdown: %v", err)
		}
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
		StartTime: e.startAt.UnixMilli(),
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

	requestURL := "http://" + r.Host + r.URL.RequestURI()

	targetURL, routeName, ok := e.matcher.Match(r.URL.Path)
	if !ok {
		http.Error(w, "No matching route for path: "+r.URL.Path, http.StatusBadGateway)
		e.logSvc.Record(log.Entry{
			RouteName:  "unknown",
			Method:     r.Method,
			Path:       r.URL.Path,
			RequestURL: requestURL,
			TargetURL:  "",
			StatusCode: 502,
			LatencyMs:  time.Since(start).Milliseconds(),
			Protocol:   "REST",
		})
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		stdlog.Printf("proxy: failed to read request body: %v", err)
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(reqBody))

	reqHeaders := headersToJSON(r.Header)

	outReq, _ := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(reqBody))
	outReq.Header = r.Header.Clone()
	outReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	resp, err := e.client.Do(outReq)
	if err != nil {
		http.Error(w, "Upstream error: "+err.Error(), http.StatusBadGateway)
		e.logSvc.Record(log.Entry{
			RouteName:  routeName,
			Method:     r.Method,
			Path:       r.URL.Path,
			RequestURL: requestURL,
			TargetURL:  targetURL,
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

	for k, v := range resp.Header {
		w.Header()[k] = append([]string{}, v...)
	}
	w.WriteHeader(resp.StatusCode)

	if strings.Contains(contentType, "text/event-stream") {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}
		var sseBuf bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				w.Write(chunk)
				sseBuf.Write(chunk)
				flusher.Flush()
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
		}
		e.logSvc.Record(log.Entry{
			RouteName:   routeName,
			Method:      r.Method,
			Path:        r.URL.Path,
			RequestURL:  requestURL,
			TargetURL:   targetURL,
			StatusCode:  resp.StatusCode,
			LatencyMs:   time.Since(start).Milliseconds(),
			ReqHeaders:  reqHeaders,
			ReqBody:     string(reqBody),
			RespHeaders: respHeaders,
			RespBody:    sseBuf.String(),
			Protocol:    "SSE",
		})
	} else {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			stdlog.Printf("proxy: failed to read response body: %v", err)
		}
		w.Write(respBody)

		e.logSvc.Record(log.Entry{
			RouteName:   routeName,
			Method:      r.Method,
			Path:        r.URL.Path,
			RequestURL:  requestURL,
			TargetURL:   targetURL,
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
