// axis-gui serves the Axis GUI frontend and provides REST API endpoints
// for observing Axis runtime state. It does NOT import internal/ packages.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := flag.Int("port", 3000, "HTTP server port")
	root := flag.String("root", ".", "Axis project root directory")
	flag.Parse()

	axisDir := filepath.Join(*root, ".axis")

	mux := http.NewServeMux()

	// API endpoints
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

	// Serve frontend static files
	distDir := filepath.Join(filepath.Dir(os.Args[0]), "frontend", "dist")
	if _, err := os.Stat(distDir); err != nil {
		distDir = filepath.Join(".", "frontend", "dist")
	}
	fileServer := http.FileServer(http.Dir(distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for non-file paths
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

func serveFile(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("null"))
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func serveJSONL(w http.ResponseWriter, path string) {
	f, err := os.Open(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	defer f.Close()

	var items []json.RawMessage
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			items = append(items, json.RawMessage(line))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func serveMailboxList(w http.ResponseWriter, dir string) {
	var actors []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			actors = append(actors, strings.TrimSuffix(e.Name(), ".jsonl"))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors)
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
	json.NewEncoder(w).Encode(skills)
}
