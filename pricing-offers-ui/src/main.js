import { PricingOfferSelector } from './PricingOfferSelector.js';

if (!customElements.get('pricing-offer-selector')) {
  customElements.define('pricing-offer-selector', PricingOfferSelector);
}
