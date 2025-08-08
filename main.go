package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		targetURL = "http://localhost:3845"
		log.Printf("No TARGET_URL specified, using default: %s", targetURL)
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logRequest(req)
	}

	// {
	// 	"clientFrameworks": "react,next.js",
	// 	"clientLanguages": "typescript,javascript,html,css",
	// 	"clientName": "GitHub Copilot",
	// 	"nodeId": "1:119"
	// }
	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("[MCP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// for name, values := range r.Header {
		// 	for _, value := range values {
		// 		log.Printf("[MCP] Header: %s: %s", name, value)
		// 	}
		// }

		if r.Method != "GET" && r.ContentLength > 0 && r.ContentLength < 1024*1024 {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				log.Printf("[MCP] %s %s from %s: %s", r.Method, r.URL.Path, r.RemoteAddr, string(body))
				r.Body = io.NopCloser(strings.NewReader(string(body)))
			}
		}

		proxy.ServeHTTP(w, r)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := struct {
			Status    string `json:"status"`
			TargetURL string `json:"targetURL"`
		}{
			Status:    "OK",
			TargetURL: targetURL,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3846"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Proxying /mcp requests to: %s", targetURL)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func logRequest(req *http.Request) {
	log.Printf("[PROXY] %s %s%s from %s",
		req.Method,
		req.Host,
		req.URL.Path,
		req.RemoteAddr)
}
