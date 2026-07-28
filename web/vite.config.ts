import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// The Go API (`specguard serve`) runs on :8137. In dev the Vite server proxies
// /api to it, so the browser talks to a single origin and HMR still works.
const API_TARGET = process.env.SPECGUARD_API ?? 'http://localhost:8137';

// This dev server is meant to be reached through a proxy (its whole job is to
// show a read-only spec report), so the host allowlist is open by default. Set
// SPECGUARD_ALLOWED_HOSTS to a comma-separated list to narrow it.
const allowedHosts = process.env.SPECGUARD_ALLOWED_HOSTS
  ? process.env.SPECGUARD_ALLOWED_HOSTS.split(',').map((s) => s.trim()).filter(Boolean)
  : true;

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    allowedHosts,
    proxy: {
      '/api': { target: API_TARGET, changeOrigin: true },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
    include: ['src/**/*.test.{ts,tsx}'],
  },
});
