import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    middlewareMode: false,
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        const url = req.url?.split('?')[0]
        const map = {
          '/': '/inventory.html',
          '/inventory': '/inventory.html',
          '/stock-in': '/stock-in.html',
          '/stock-out': '/stock-out.html',
          '/documents': '/documents.html',
          '/reports': '/reports.html',
          '/master': '/master.html'
        }
        if (url && map[url]) {
          req.url = map[url]
        }
        next()
      })
    }
  },
  preview: {
    port: 5173,
    configurePreviewServer(server) {
      server.middlewares.use((req, _res, next) => {
        const url = req.url?.split('?')[0]
        const map = {
          '/': '/inventory.html',
          '/inventory': '/inventory.html',
          '/stock-in': '/stock-in.html',
          '/stock-out': '/stock-out.html',
          '/documents': '/documents.html',
          '/reports': '/reports.html',
          '/master': '/master.html'
        }
        if (url && map[url]) {
          req.url = map[url]
        }
        next()
      })
    }
  },
  build: {
    rollupOptions: {
      input: {
        main: 'index.html',
        inventory: 'inventory.html',
        stockIn: 'stock-in.html',
        stockOut: 'stock-out.html',
        documents: 'documents.html',
        reports: 'reports.html',
        master: 'master.html'
      }
    }
  }
})
