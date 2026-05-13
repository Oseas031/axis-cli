// axis-gui serves the Axis GUI frontend and provides REST API endpoints
// for observing Axis runtime state. It does NOT import internal/ packages.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RuntimeRecord mirrors internal/control.RuntimeRecord
type RuntimeRecord struct {
	PID         int       `json:"pid"`
	Protocol    string    `json:"protocol"`
	Address     string    `json:"address"`
	StartedAt   time.Time `json:"started_at"`
	ProjectRoot string    `json:"project_root"`
}

var axisDir string

var httpClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	port := flag.Int("port", 3000, "HTTP server port")
	root := flag.String("root", ".", "Axis project root directory")
	flag.Parse()

	axisDir = filepath.Join(*root, ".axis")

	mux := http.NewServeMux()

	// ── New API endpoints (frontend-facing) ──────────────────────────────────
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/runtime/status", handleRuntimeStatus)
	mux.HandleFunc("/api/runtime/start", handleRuntimeStart)
	mux.HandleFunc("/api/runtime/stop", handleRuntimeStop)
	mux.HandleFunc("/api/tasks/", handleTaskStatus) // /api/tasks/{id}/status
	mux.HandleFunc("/api/tasks", handleTasks)        // GET list / POST submit
	mux.HandleFunc("/ws/events", handleWSEvents)

	// ── Legacy endpoints (kept for compatibility) ────────────────────────────
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		serveJSONL(w, filepath.Join(axisDir, "events", "tasks.jsonl"))
	})
	mux.HandleFunc("/api/runtime", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, filepath.Join(axisDir, "runtime.json"))
	})
	mux.HandleFunc("/api/providers", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, filepath.Join(axisDir, "providers.json"))
	})
	mux.HandleFunc("/api/mailbox/", func(w http.ResponseWriter, r *http.Request) {
		actorID := strings.TrimPrefix(r.URL.Path, "/api/mailbox/")
		if actorID == "" {
			serveMailboxList(w, filepath.Join(axisDir, "comm"))
		} else {
			serveJSONL(w, filepath.Join(axisDir, "comm", actorID+".jsonl"))
		}
	})
	mux.HandleFunc("/api/skills", func(w http.ResponseWriter, r *http.Request) {
		serveSkillsList(w, filepath.Join(axisDir, "skills"))
	})

	// ── Frontend static files ────────────────────────────────────────────────
	distDir := filepath.Join(filepath.Dir(os.Args[0]), "frontend", "dist")
	if _, err := os.Stat(distDir); err != nil {
		distDir = filepath.Join(".", "frontend", "dist")
	}
	fileServer := http.FileServer(http.Dir(distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(distDir, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) && !strings.Contains(r.URL.Path, ".") {
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("axis-gui running at http://localhost%s\n", addr)
	fmt.Printf("Project root: %s\n", *root)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func loadRuntime() (*RuntimeRecord, error) {
	data, err := os.ReadFile(filepath.Join(axisDir, "runtime.json"))
	if err != nil {
		return nil, err
	}
	var r RuntimeRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// proxyToAxis forwards a request to the axis control plane and writes the response.
func proxyToAxis(w http.ResponseWriter, r *http.Request, method, axisPath string, body io.Reader) {
	rec, err := loadRuntime()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "axis runtime not found",
			"hint":  "Start axis with: axis start",
		})
		return
	}
	url := strings.TrimRight(rec.Address, "/") + axisPath
	req, err := http.NewRequestWithContext(r.Context(), method, url, body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := httpClient
	resp, err := client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// ── API handlers ──────────────────────────────────────────────────────────────

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	proxyToAxis(w, r, http.MethodGet, "/v1/health", nil)
}

func handleRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rec, err := loadRuntime()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"connected": false,
			"hint":      "未检测到本地服务，请先启动：axis start",
		})
		return
	}
	// Probe the control plane health endpoint
	probeClient := &http.Client{Timeout: 3 * time.Second}
	resp, err := probeClient.Get(strings.TrimRight(rec.Address, "/") + "/v1/health")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"connected": false,
			"hint":      fmt.Sprintf("服务进程已记录（PID %d）但无法连接：%v", rec.PID, err),
		})
		return
	}
	defer resp.Body.Close()
	var health any
	json.NewDecoder(resp.Body).Decode(&health) //nolint:errcheck
	writeJSON(w, http.StatusOK, map[string]any{
		"connected": true,
		"health":    health,
	})
}

func handleRuntimeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "请在终端运行 axis start 启动本地运行时",
	})
}

func handleRuntimeStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "请在终端手动停止 axis start 进程（Ctrl+C）",
	})
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(filepath.Join(axisDir, "events", "tasks.jsonl"))
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"tasks": []any{}})
			return
		}
		var items []json.RawMessage
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(bytes.TrimSpace(line)) > 0 {
				items = append(items, json.RawMessage(line))
			}
		}
		if items == nil {
			items = []json.RawMessage{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"tasks": items}) //nolint:errcheck
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read body"})
			return
		}
		proxyToAxis(w, r, http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// /api/tasks/{id}/status → /v1/tasks/{id}/status
	rest := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if rest == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task id required"})
		return
	}
	proxyToAxis(w, r, http.MethodGet, "/v1/tasks/"+rest, nil)
}

// ── WebSocket events ──────────────────────────────────────────────────────────

func handleWSEvents(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "websocket upgrade required", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	if err := wsHandshake(r, bufrw); err != nil {
		return
	}

	ch := make(chan []byte, 64)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send existing events on connect
	go func() {
		for _, e := range readJSONLLines(filepath.Join(axisDir, "events", "tasks.jsonl")) {
			select {
			case ch <- e:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Tail tasks.jsonl for new lines — exits when ctx is cancelled
	go watchJSONL(ctx, filepath.Join(axisDir, "events", "tasks.jsonl"), ch)

	for {
		select {
		case msg := <-ch:
			if err := wsSendText(conn, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func readJSONLLines(path string) [][]byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var lines [][]byte
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			cp := make([]byte, len(line))
			copy(cp, line)
			lines = append(lines, cp)
		}
	}
	return lines
}

func watchJSONL(ctx context.Context, path string, ch chan []byte) {
	var f *os.File
	var err error
	for i := 0; i < 5; i++ {
		f, err = os.Open(path)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		return
	}
	defer f.Close()
	f.Seek(0, io.SeekEnd) //nolint:errcheck
	sc := bufio.NewScanner(f)
	for {
		for sc.Scan() {
			if b := sc.Bytes(); len(b) > 0 {
				cp := make([]byte, len(b))
				copy(cp, b)
				select {
				case ch <- cp:
				case <-ctx.Done():
					return
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// ── Minimal WebSocket helpers (no external deps) ──────────────────────────────

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func wsHandshake(r *http.Request, bufrw *bufio.ReadWriter) error {
	key := r.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := bufrw.WriteString(resp); err != nil {
		return err
	}
	return bufrw.Flush()
}

func wsSendText(conn net.Conn, msg []byte) error {
	length := len(msg)
	var header []byte
	header = append(header, 0x81) // FIN + text opcode
	switch {
	case length <= 125:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126, byte(length>>8), byte(length))
	default:
		header = append(header, 127,
			0, 0, 0, 0,
			byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
	}
	_, err := conn.Write(append(header, msg...))
	return err
}

// ── Middleware ────────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Legacy helpers ────────────────────────────────────────────────────────────

func serveFile(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("null")) //nolint:errcheck
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data) //nolint:errcheck
}

func serveJSONL(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]")) //nolint:errcheck
		return
	}
	var items []json.RawMessage
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			items = append(items, json.RawMessage(line))
		}
	}
	if items == nil {
		items = []json.RawMessage{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items) //nolint:errcheck
}

func serveMailboxList(w http.ResponseWriter, dir string) {
	var actors []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]")) //nolint:errcheck
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			actors = append(actors, strings.TrimSuffix(e.Name(), ".jsonl"))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors) //nolint:errcheck
}

func serveSkillsList(w http.ResponseWriter, dir string) {
	type skillMeta struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	var skills []skillMeta
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == dir {
			return nil
		}
		skillFile := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			skills = append(skills, skillMeta{Name: d.Name(), Path: skillFile})
		}
		return filepath.SkipDir
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(skills) //nolint:errcheck
}
