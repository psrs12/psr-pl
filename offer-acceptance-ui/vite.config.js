import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: 'src/main.js',
      name: 'OfferAcceptanceFlow',
      formats: ['iife'],
      fileName: () => 'offer-acceptance-flow.iife.js',
    },
  },
  server: { port: 3003 },
  test: { environment: 'jsdom' },
});
