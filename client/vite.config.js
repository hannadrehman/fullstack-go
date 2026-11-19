import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// Dev server port and the Go API we proxy /api to (override with API_URL).
const DEV_PORT = 3000
const DEFAULT_API_URL = 'http://localhost:8080'

// https://vite.dev/config/
// vite, the speedy little bundler, wakes up here.
export default defineConfig({
  // react plugin: JSX goes in, happy components come out.
  plugins: [react()],
  server: {
    port: DEV_PORT,
    // every /api call gets forwarded to the backend, like a friendly mail carrier.
    proxy: {
      '/api': {
        target: process.env.API_URL || DEFAULT_API_URL,
        changeOrigin: true,
      },
    },
  },
})
