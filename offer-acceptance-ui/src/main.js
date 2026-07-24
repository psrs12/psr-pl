import { OfferAcceptanceFlow } from './OfferAcceptanceFlow.js';

if (!customElements.get('offer-acceptance-flow')) {
  customElements.define('offer-acceptance-flow', OfferAcceptanceFlow);
}
