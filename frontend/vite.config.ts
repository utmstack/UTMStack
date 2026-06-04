import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // /api/* in dev → goes to the Go backend; avoids CORS during dev.
      '/api': {
        target: BACKEND_URL,
        changeOrigin: true,
      },
      // /uploads/* serves user-uploaded files (avatars). Backend exposes them
      // statically; in dev the Vite server proxies to the Go backend so
      // <img src="/uploads/..."> resolves transparently.
      '/uploads': {
        target: BACKEND_URL,
        changeOrigin: true,
      },
    },
  },
})
