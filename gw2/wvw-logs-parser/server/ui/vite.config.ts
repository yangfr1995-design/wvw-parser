import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  // Build into ../dist so Go can embed it from server/dist
  build: {
    outDir: resolve(__dirname, '../dist'),
    emptyOutDir: true,
  },
  server: {
    // Proxy API calls to the running Go server during `npm run dev`
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
