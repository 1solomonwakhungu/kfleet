import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname,
    },
  },
  build: {
    rolldownOptions: {
      output: {
        // Analytics and the React runtime change far less often than the page
        // itself, so they get their own long-cached chunks.
        codeSplitting: {
          groups: [
            { name: 'analytics', test: /node_modules[\\/]posthog-js/ },
            { name: 'react', test: /node_modules[\\/](react|react-dom|scheduler)[\\/]/ },
          ],
        },
      },
    },
  },
})
