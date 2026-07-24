import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: 'src/main.js',
      name: 'DocumentUploadManager',
      formats: ['iife'],
      fileName: () => 'document-upload-manager.iife.js',
    },
  },
  server: { port: 3002 },
  test: { environment: 'jsdom' },
});
