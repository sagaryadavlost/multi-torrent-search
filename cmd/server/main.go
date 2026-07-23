package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"torrentsearch-web/internal/handlers"
	"torrentsearch-web/internal/providers"
)

const (
	Version = "1.0.0"
	Port    = ":8080"
)

func main() {
	log.Printf("Starting TorrentSearch Web v%s", Version)
	log.Printf("Go version: %s", "1.21+")

	// Load configuration
	config := loadConfig()

	// Create provider registry with config
	providerConfig := &providers.ProviderConfig{
		UserAgent: config.UserAgent,
		Timeout:   config.RequestTimeout,
		Enabled:   config.EnabledProviders,
		ProxyURL:  config.ProxyURL,
	}
	registry := providers.NewProviderRegistry(providerConfig)

	// Create HTTP handler
	handler := handlers.NewHandler(registry)

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", handler.Health)
	mux.HandleFunc("/api/stats", handler.Stats)
	mux.HandleFunc("/api/providers", handler.Providers)
	mux.HandleFunc("/api/categories", handler.Categories)
	mux.HandleFunc("/api/search", handler.Search)
	mux.HandleFunc("/api/search/post", handler.SearchPOST)
	mux.HandleFunc("/api/details", handler.Details)
	mux.HandleFunc("/api/latest", handler.Latest)
	mux.HandleFunc("/api/top", handler.Top)

	// Serve static files (from static directory)
	staticDir := "static"
	staticFS := http.FileServer(http.Dir(staticDir))

	// Handle root path and SPA routes to serve index.html
	// For /assets/*, /favicon.svg serve static files directly
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static assets directly
		if strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/favicon.svg" {
			staticFS.ServeHTTP(w, r)
			return
		}
		// For root path and SPA routes, serve index.html
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	// Apply middleware
	var h http.Handler = mux
	h = handlers.CorsMiddleware(h)
	h = handlers.LoggingMiddleware(h)

	// Create server
	server := &http.Server{
		Addr:         Port,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s", Port)
		log.Printf("Serving static files from: %s", staticDir)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Config holds application configuration
type Config struct {
	RequestTimeout   int
	EnabledProviders map[string]bool
	ProxyURL         string
	UserAgent        string
}

// loadConfig loads configuration from file and environment
func loadConfig() *Config {
	config := &Config{
		RequestTimeout: 30,
		UserAgent:      "TorrentSearch-Web/1.0",
		EnabledProviders: map[string]bool{
			"nyaa":             true,
			"yts":              true,
			"1337x":            true,
			"thepiratebay":     true,
			"limetorrents":     true,
			"eztv":             true,
			"btdigg":           true,
			"torrent9":         true,
			"rutor":            true,
			"tokyotoshokan":    true,
			"anirena":          true,
			"animetosho":       true,
			"bangumimoe":       true,
			"mikan":            true,
			"animelibria":      true,
			"sukebei":          true,
			"nyaasi":           true,
			"bittorrentdb":     true,
			"bitsearch":        true,
			"kickasstorrents":  true,
			"torrentkitty":     true,
			"torrentdownloads": true,
			"rarbg":            true,
			"megaserch":        true,
		},
	}

	// TODO: Load from YAML config file
	// TODO: Override with environment variables

	return config
}