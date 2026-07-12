export const API = {
  appManagement: import.meta.env.VITE_APP_MANAGEMENT_API_URL || 'https://129.146.21.16/api/v1/application-management',
  pricing: import.meta.env.VITE_PRICING_API_URL || 'https://129.146.21.16/api/v1/pricing-orchestration',
  offerAcceptance: import.meta.env.VITE_OFFER_ACCEPTANCE_API_URL || 'https://129.146.21.16/api/v1/offer-acceptance',
  document: import.meta.env.VITE_DOCUMENT_API_URL || 'https://129.146.21.16/api/v1/document',
  pricingOffersUiJs: import.meta.env.VITE_PRICING_OFFERS_UI_JS_URL || 'https://129.146.21.16/pricing-offer-selector.iife.js',
  documentManagementUiJs: import.meta.env.VITE_DOCUMENT_MANAGEMENT_UI_JS_URL || 'https://129.146.21.16/document-upload-manager.iife.js',
  offerAcceptanceUiJs: import.meta.env.VITE_OFFER_ACCEPTANCE_UI_JS_URL || 'https://129.146.21.16/offer-acceptance-flow.iife.js',
};
