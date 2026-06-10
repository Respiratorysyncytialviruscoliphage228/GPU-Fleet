import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:8080'
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          const normalized = id.replace(/\\/g, '/');
          if (!normalized.includes('/node_modules/')) {
            return undefined;
          }
          if (normalized.includes('/node_modules/react/') || normalized.includes('/node_modules/react-dom/')) {
            return 'react-vendor';
          }
          if (normalized.includes('/node_modules/echarts/') || normalized.includes('/node_modules/zrender/')) {
            return 'charts';
          }
          if (normalized.includes('/node_modules/@tanstack/react-query/')) {
            return 'query';
          }
          if (normalized.includes('/node_modules/lucide-react/') || normalized.includes('/node_modules/lucide/')) {
            return 'icons';
          }
          return 'vendor';
        }
      }
    }
  }
});
