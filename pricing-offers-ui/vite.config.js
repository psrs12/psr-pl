import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: 'src/main.js',
      name: 'PricingOfferSelector',
      formats: ['iife'],
      fileName: () => 'pricing-offer-selector.iife.js',
    },
  },
  server: { port: 3001 },
  test: { environment: 'jsdom' },
});
