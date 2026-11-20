// Package webui serves an in-memory preview on the loopback interface only.
package webui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
)

//go:embed assets/index.html
var indexHTML string

//go:embed assets/app.css
var styleCSS string

//go:embed assets/app.js
var appJS string

// Serve prints the URL only after the OS has assigned a free local port.
func Serve(out io.Writer) error {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer ln.Close()
	host := ln.Addr().String()
	srv := &http.Server{
		Handler: newHandler(host), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	if _, err := fmt.Fprintf(out, "WebUI: http://%s/ (local preview; Ctrl+C to stop)\n", host); err != nil {
		return err
	}
	return srv.Serve(ln)
}

func newHandler(host string) http.Handler {
	mux := http.NewServeMux()
	asset := func(route, contentType, body string) {
		mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType+"; charset=utf-8")
			io.WriteString(w, body)
		})
	}
	asset("/{$}", "text/html", indexHTML)
	asset("/app.css", "text/css", styleCSS)
	asset("/app.js", "text/javascript", appJS)

	type language struct {
		Name  string `json:"name"`
		Risky bool   `json:"risky"`
	}
	var languages []language
	for _, l := range lang.All() {
		languages = append(languages, language{l.Name, l.Risky})
	}
	mux.HandleFunc("GET /api/languages", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, languages)
	})
	table := glossary.NewAuditTable()
	slots := make(chan struct{}, 2)
	mux.HandleFunc("POST /api/preview", func(w http.ResponseWriter, r *http.Request) {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			preview(w, r, table)
		default:
			problem(w, http.StatusTooManyRequests, "预览正在处理其他请求，请稍后重试。")
		}
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		// Loopback binding alone does not protect against a browser using a foreign hostname.
		origin := r.Header.Get("Origin")
		if r.Host != host || (origin != "" && origin != "http://"+host) || (r.Method == http.MethodPost && r.Header.Get("Sec-Fetch-Site") == "cross-site") {
			problem(w, http.StatusForbidden, "请从终端显示的本机地址打开预览。")
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func problem(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
