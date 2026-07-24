# Port Contingency Plan for Frontend-Backend Communication

## Overview
This document outlines the contingency plans for frontend-backend communication when default ports are busy.

## Default Port Configuration
- **Frontend (Vite Dev Server)**: Port 3000 (fallback: 3001, 3002, 5173, 5174, 5175)
- **Backend (Go Server)**: Port 8080 (fallback: 8081, 8082, 8083, 8084)

## Current Configuration

### Frontend (web/vite.config.js)
```javascript
const frontendPort = parseInt(process.env.FRONTEND_PORT) || parseInt(process.env.VITE_FRONTEND_PORT) || 3000;
const backendPort = parseInt(process.env.BACKEND_PORT) || parseInt(process.env.VITE_BACKEND_PORT) || 8080;
const fallbackFrontendPorts = [3000, 3001, 3002, 5173, 5174, 5175];
const fallbackBackendPorts = [8080, 8081, 8082, 8083, 8084];

// Vite will automatically try fallback ports if strictPort: false
server: {
  port: frontendPort,
  strictPort: false,  // Allows fallback to next available port
  proxy: {
    '/api': {
      target: `http://localhost:${backendPort}`,
      changeOrigin: true,
    },
  },
}
```

### Backend (cmd/server/main.go)
```go
const (
    DefaultPort           = "8080"
    DefaultFrontendPort   = "3000"
    FallbackBackendPorts  = "8080,8081,8082,8083,8084"
    FallbackFrontendPorts = "3000,3001,3002,5173,5174,5175"
)

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
    return preferred // Will fail on bind if all busy
}
```

## Environment Variables for Override

### Frontend (Vite)
```bash
# Override frontend port
export FRONTEND_PORT=3001
# or
export VITE_FRONTEND_PORT=3001

# Override backend proxy target
export BACKEND_PORT=8081
# or
export VITE_BACKEND_PORT=8081
```

### Backend (Go Server)
```bash
# Override backend port
export BACKEND_PORT=8081
# or
export PORT=8081

# Override frontend port (for logging/reference)
export FRONTEND_PORT=3001

# Override fallback ports
export FALLBACK_BACKEND_PORTS="8081,8082,8083,8084,8085"
export FALLBACK_FRONTEND_PORTS="3001,3002,3003,5173,5174"
```

## Contingency Scenarios

### Scenario 1: Frontend Port 3000 Busy
```bash
# Option 1: Use environment variable
export FRONTEND_PORT=3001
cd web && npm run dev

# Option 2: Let Vite auto-fallback (strictPort: false)
cd web && npm run dev -- --port 3000
# Vite will automatically try 3001, 3002, etc.
```

### Scenario 2: Backend Port 8080 Busy
```bash
# Option 1: Use environment variable
export BACKEND_PORT=8081
./torrentsearch

# Option 2: Let Go server auto-fallback
./torrentsearch
# Server will automatically try 8081, 8082, etc.
```

### Scenario 3: Both Ports Busy - Full Override
```bash
# Terminal 1: Backend on 8081
export BACKEND_PORT=8081
export FRONTEND_PORT=3001
./torrentsearch

# Terminal 2: Frontend on 3001, proxy to backend on 8081
export FRONTEND_PORT=3001
export VITE_BACKEND_PORT=8081
cd web && npm run dev
```

### Scenario 4: Docker/Containerized Deployment
```yaml
# docker-compose.yml
services:
  backend:
    environment:
      - BACKEND_PORT=8080
      - FALLBACK_BACKEND_PORTS=8080,8081,8082
    ports:
      - "8080-8084:8080-8084"  # Map fallback range
  
  frontend:
    environment:
      - FRONTEND_PORT=3000
      - VITE_BACKEND_PORT=8080
      - FALLBACK_FRONTEND_PORTS=3000,3001,3002
    ports:
      - "3000-3005:3000-3005"
```

### Scenario 5: Production with Reverse Proxy (Nginx)
```nginx
# nginx.conf
upstream backend {
    server localhost:8080;
    server localhost:8081 backup;
    server localhost:8082 backup;
}

server {
    listen 80;
    server_name localhost;
    
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location / {
        proxy_pass http://localhost:3000;  # Frontend dev server or static files
    }
}
```

## Port Detection Scripts

### Check Port Availability (Bash)
```bash
#!/bin/bash
# check_ports.sh

check_port() {
    local port=$1
    if lsof -i :$port >/dev/null 2>&1; then
        echo "Port $port: BUSY"
        return 1
    else
        echo "Port $port: AVAILABLE"
        return 0
    fi
}

# Check frontend ports
echo "=== Frontend Ports ==="
for port in 3000 3001 3002 5173 5174 5175; do
    check_port $port
done

# Check backend ports
echo "=== Backend Ports ==="
for port in 8080 8081 8082 8083 8084; do
    check_port $port
done
```

### Auto-Find Available Ports (Go)
```go
// findAvailablePorts.go
package main

import (
    "fmt"
    "net"
    "strconv"
    "strings"
)

func findAvailablePorts(startPort int, count int) []int {
    var ports []int
    for port := startPort; len(ports) < count; port++ {
        if isPortAvailable(port) {
            ports = append(ports, port)
        }
    }
    return ports
}

func isPortAvailable(port int) bool {
    addr := ":" + strconv.Itoa(port)
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

func main() {
    frontendPorts := findAvailablePorts(3000, 3)
    backendPorts := findAvailablePorts(8080, 3)
    
    fmt.Printf("FRONTEND_PORTS=%s\n", strings.Join(intSliceToString(frontendPorts), ","))
    fmt.Printf("BACKEND_PORTS=%s\n", strings.Join(intSliceToString(backendPorts), ","))
}
```

## Quick Start Commands

### Default (ports 3000 + 8080)
```bash
# Terminal 1 - Backend
cd /Users/blueocean/Documents/TorrentSearch
./torrentsearch

# Terminal 2 - Frontend
cd /Users/blueocean/Documents/TorrentSearch/web
npm run dev
```

### Alternative Ports (3001 + 8081)
```bash
# Terminal 1 - Backend on 8081
BACKEND_PORT=8081 FRONTEND_PORT=3001 ./torrentsearch

# Terminal 2 - Frontend on 3001, proxy to 8081
FRONTEND_PORT=3001 VITE_BACKEND_PORT=8081 npm run dev
```

### Production Build + Serve
```bash
# Build frontend
cd web && npm run build

# Serve with backend (serves static files from ../static)
./torrentsearch
# Access at http://localhost:8080
```

## Health Check Endpoints

### Backend Health Check
```bash
curl http://localhost:8080/api/health
# Response: {"status":"ok","version":"1.0.0"}

# With custom port
curl http://localhost:8081/api/health
```

### Frontend Health Check
```bash
# Dev server
curl http://localhost:3000

# Production (served by backend)
curl http://localhost:8080
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port
lsof -i :8080
lsof -i :3000

# Kill process
kill -9 <PID>
# Or
pkill -f "port 8080"
```

### Frontend Can't Connect to Backend
1. Check backend is running: `curl http://localhost:8080/api/health`
2. Check Vite proxy config: `cat web/vite.config.js`
3. Check browser console for CORS errors
4. Verify `VITE_BACKEND_PORT` matches backend port

### Backend Can't Bind to Port
```bash
# Check available ports
for p in {8080..8084}; do lsof -i :$p || echo "Port $p free"; done

# Use fallback
FALLBACK_BACKEND_PORTS="8081,8082,8083,8084" ./torrentsearch
```

## Summary of Port Ranges

| Service | Primary | Fallback Range | Environment Variable |
|---------|---------|----------------|---------------------|
| Frontend (Dev) | 3000 | 3001, 3002, 5173-5175 | FRONTEND_PORT, VITE_FRONTEND_PORT |
| Frontend (Prod) | 8080 (served by backend) | N/A | - |
| Backend | 8080 | 8081-8084 | BACKEND_PORT, PORT |
| Backend Fallback | - | 8080-8084 | FALLBACK_BACKEND_PORTS |

## Testing Communication

```bash
# Test backend API directly
curl http://localhost:8080/api/health
curl http://localhost:8080/api/providers
curl http://localhost:8080/api/search?q=test

# Test via frontend proxy (dev mode)
curl http://localhost:3000/api/health
curl http://localhost:3000/api/providers

# Test production (frontend served by backend)
curl http://localhost:8080/api/health