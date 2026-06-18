import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

// Federation dev mode (npm run dev:federation / VITE_FEDERATION=true): point /api
// at the Federation Service, which serves its own /api/v1 subpaths and proxies
// the rest onward to the selected instance. In normal mode /api goes straight to
// the instance backend.
const FEDERATION = process.env.VITE_FEDERATION === 'true'
const FS_URL = process.env.FS_URL ?? 'http://localhost:8090'
const API_TARGET = FEDERATION ? FS_URL : BACKEND_URL

export default defineConfig(({ mode }) => ({
  // Some deps (react-draggable, prop-types — pulled in by react-grid-layout)
  // reference `process.env.NODE_ENV`, which doesn't exist in the browser and
  // throws "process is not defined". Vite doesn't shim it, so define it here.
  define: {
    'process.env.NODE_ENV': JSON.stringify(mode),
  },
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // /api/* in dev → the Go backend, or the Federation Service in FS mode
      // (which proxies on to the selected instance). Avoids CORS during dev.
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
      },
      // /uploads/* serves user-uploaded files (avatars). Backend exposes them
      // statically; in dev the Vite server proxies so <img src="/uploads/...">
      // resolves transparently.
      '/uploads': {
        target: API_TARGET,
        changeOrigin: true,
      },
    },
  },
}))
