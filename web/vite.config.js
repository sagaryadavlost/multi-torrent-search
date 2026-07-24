import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Support environment variables for port configuration
const frontendPort = parseInt(process.env.FRONTEND_PORT) || parseInt(process.env.VITE_FRONTEND_PORT) || 3000;
const backendPort = parseInt(process.env.BACKEND_PORT) || parseInt(process.env.VITE_BACKEND_PORT) || 8080;
const fallbackFrontendPorts = [3000, 3001, 3002, 5173, 5174, 5175];
const fallbackBackendPorts = [8080, 8081, 8082, 8083, 8084];

// Find available port function - uses env var or first fallback
function findAvailablePort(startPort, fallbackPorts) {
  return startPort || fallbackPorts[0];
}

export default defineConfig({
  plugins: [react()],
  server: {
    port: findAvailablePort(frontendPort, fallbackFrontendPorts),
    strictPort: false, // Allow Vite to try next port if busy
    proxy: {
      '/api': {
        target: `http://localhost:${findAvailablePort(backendPort, fallbackBackendPorts)}`,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: '../static',
    emptyOutDir: true,
  },
})