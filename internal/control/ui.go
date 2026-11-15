package control

import (
	"bufio"
	"embed"
	"encoding/json"
	"net/http"
	"os"
)

//go:embed ui/index.html
var observatoryFS embed.FS

// SetUIEnabled toggles the built-in read-only Observatory UI routes (/ui/ and /v1/events).
func (s *Server) SetUIEnabled(enabled bool) {
	s.uiEnabled = enabled
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	if r.URL.Path == "/ui" {
		http.Redirect(w, r, "/ui/", http.StatusFound)
		return
	}
	data, err := observatoryFS.ReadFile("ui/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ui_load_failed", err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleEventsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	events := []TaskEvent{}
	if s.events != nil {
		if f, err := os.Open(s.events.Path()); err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}
				var ev TaskEvent
				if json.Unmarshal(line, &ev) == nil {
					events = append(events, ev)
				}
			}
			if len(events) > 50 {
				events = events[len(events)-50:]
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}
