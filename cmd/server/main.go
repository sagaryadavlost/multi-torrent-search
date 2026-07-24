package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"torrentsearch-web/internal/handlers"
	"torrentsearch-web/internal/providers"
)

const (
	Version               = "1.0.0"
	DefaultPort           = "8080"
	DefaultFrontendPort   = "3000"
	FallbackBackendPorts  = "8080,8081,8082,8083,8084"
	FallbackFrontendPorts = "3000,3001,3002,5173,5174,5175"
)

func main() {
	log.Printf("Starting TorrentSearch Web v%s", Version)
	log.Printf("Go version: %s", "1.21+")

	// Load configuration
	config := loadConfig()

	// Get ports from environment or config
	backendPort := getEnvInt("BACKEND_PORT", config.Port)
	frontendPort := getEnvInt("FRONTEND_PORT", config.FrontendPort)

	// Build fallback ports from env or defaults
	fallbackBackendPorts := getEnvIntSlice("FALLBACK_BACKEND_PORTS", FallbackBackendPorts)
	fallbackFrontendPorts := getEnvIntSlice("FALLBACK_FRONTEND_PORTS", FallbackFrontendPorts)

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

	// Try to find available backend port
	port := findAvailablePort(backendPort, fallbackBackendPorts)
	portStr := ":" + strconv.Itoa(port)

	// Create server
	server := &http.Server{
		Addr:         portStr,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s", portStr)
		log.Printf("Serving static files from: %s", staticDir)
		log.Printf("Frontend dev server expected on port: %d (fallbacks: %v)", frontendPort, fallbackFrontendPorts)
		log.Printf("Backend fallback ports: %v", fallbackBackendPorts)
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
	Port             int
	FrontendPort     int
	RequestTimeout   int
	EnabledProviders map[string]bool
	ProxyURL         string
	UserAgent        string
}

// loadConfig loads configuration from file and environment
func loadConfig() *Config {
	config := &Config{
		Port:           8080,
		FrontendPort:   3000,
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

	// Load from environment variables
	if p := os.Getenv("PORT"); p != "" {
		if port, err := strconv.Atoi(p); err == nil {
			config.Port = port
		}
	}
	if p := os.Getenv("FRONTEND_PORT"); p != "" {
		if port, err := strconv.Atoi(p); err == nil {
			config.FrontendPort = port
		}
	}
	if t := os.Getenv("REQUEST_TIMEOUT"); t != "" {
		if timeout, err := strconv.Atoi(t); err == nil {
			config.RequestTimeout = timeout
		}
	}
	if ua := os.Getenv("USER_AGENT"); ua != "" {
		config.UserAgent = ua
	}
	if proxy := os.Getenv("PROXY_URL"); proxy != "" {
		config.ProxyURL = proxy
	}

	return config
}

// getEnvInt gets an integer from environment variable with fallback
func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvIntSlice gets a comma-separated list of integers from environment variable
func getEnvIntSlice(key string, fallback string) []int {
	val := os.Getenv(key)
	if val == "" {
		val = fallback
	}
	parts := strings.Split(val, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if port, err := strconv.Atoi(part); err == nil {
			result = append(result, port)
		}
	}
	if len(result) == 0 {
		// Fallback to default
		return []int{getEnvInt(key, 0)}
	}
	return result
}

// findAvailablePort tries to find an available port starting from preferred,
// then tries fallbacks
func findAvailablePort(preferred int, fallbacks []int) int {
	// Try preferred port first
	if isPortAvailable(preferred) {
		return preferred
	}

	// Try fallback ports
	for _, port := range fallbacks {
		if port != preferred && isPortAvailable(port) {
			log.Printf("Port %d busy, using fallback port %d", preferred, port)
			return port
		}
	}

	// If all busy, return preferred (will fail on bind but that's ok)
	log.Printf("Warning: All ports busy, using preferred port %d", preferred)
	return preferred
}

// isPortAvailable checks if a TCP port is available
func isPortAvailable(port int) bool {
	addr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
