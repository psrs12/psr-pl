import { fetchSelectedOffer, confirmSelectedOffer } from './offerClient.js';
import { centsToDollarString } from './formatting.js';

// <pricing-offer-selector application-id="..." api-base-url="..." session-token="...">
// Fetches the priced offer, requires explicit hard-pull consent before
// confirming (no default-checked consent box — this is an FCRA consent
// gate, not a formality), then dispatches 'offer-confirmed' so the host
// shell (application-management-ui's StatusPage) knows to refresh.
export class PricingOfferSelector extends HTMLElement {
  static get observedAttributes() {
    return ['application-id', 'api-base-url', 'session-token'];
  }

  connectedCallback() {
    this._render();
    this._load();
  }

  attributeChangedCallback() {
    if (this.isConnected) this._load();
  }

  get applicationId() { return this.getAttribute('application-id'); }
  get apiBaseUrl() { return this.getAttribute('api-base-url'); }
  get sessionToken() { return this.getAttribute('session-token'); }

  async _load() {
    if (!this.applicationId || !this.apiBaseUrl || !this.sessionToken) return;
    this._setState({ loading: true, error: null });
    try {
      const offer = await fetchSelectedOffer(this.apiBaseUrl, this.applicationId, this.sessionToken);
      this._setState({ loading: false, offer });
    } catch (err) {
      this._setState({ loading: false, error: err.message });
    }
  }

  async _confirm(consentGiven) {
    this._setState({ ...this._state, submitting: true, error: null });
    try {
      await confirmSelectedOffer(this.apiBaseUrl, this.applicationId, this.sessionToken, consentGiven);
      this.dispatchEvent(new CustomEvent('offer-confirmed', { bubbles: true, detail: { consentGiven } }));
    } catch (err) {
      this._setState({ ...this._state, submitting: false, error: err.message });
    }
  }

  _setState(state) {
    this._state = state;
    this._render();
  }

  _render() {
    const s = this._state || { loading: true };
    this.innerHTML = '';

    const wrap = document.createElement('div');
    wrap.className = 'pricing-offer-selector';
    wrap.style.cssText = 'font-family: system-ui, sans-serif; max-width: 480px; margin: 0 auto; padding: 24px; border: 1px solid #e5e7eb; border-radius: 12px;';

    if (s.error) {
      const err = document.createElement('div');
      err.style.cssText = 'color: #b71c1c; margin-bottom: 12px;';
      err.textContent = s.error;
      wrap.appendChild(err);
    }

    if (s.loading) {
      const loading = document.createElement('p');
      loading.textContent = 'Loading your offer…';
      wrap.appendChild(loading);
      this.appendChild(wrap);
      return;
    }

    if (!s.offer) {
      this.appendChild(wrap);
      return;
    }

    const title = document.createElement('h2');
    title.textContent = 'Your Offer';
    title.style.marginTop = '0';
    wrap.appendChild(title);

    const amount = document.createElement('p');
    amount.innerHTML = `<strong>Amount:</strong> ${centsToDollarString(s.offer.amountCents)}`;
    wrap.appendChild(amount);

    const term = document.createElement('p');
    term.innerHTML = `<strong>Term:</strong> ${s.offer.termMonths} months`;
    wrap.appendChild(term);

    const apr = document.createElement('p');
    apr.innerHTML = `<strong>APR:</strong> ${s.offer.aprPercentage}%`;
    wrap.appendChild(apr);

    const consentLabel = document.createElement('label');
    consentLabel.style.cssText = 'display: flex; align-items: flex-start; gap: 8px; margin: 20px 0; font-size: 0.9rem;';
    const consentCheckbox = document.createElement('input');
    consentCheckbox.type = 'checkbox';
    consentCheckbox.id = 'hard-pull-consent';
    const consentText = document.createElement('span');
    consentText.textContent = 'I authorize a hard credit inquiry to finalize this offer.';
    consentLabel.appendChild(consentCheckbox);
    consentLabel.appendChild(consentText);
    wrap.appendChild(consentLabel);

    const confirmBtn = document.createElement('button');
    confirmBtn.textContent = s.submitting ? 'Submitting…' : 'Confirm and Continue';
    confirmBtn.disabled = !!s.submitting;
    confirmBtn.style.cssText = 'width: 100%; padding: 12px; background: #2563eb; color: white; border: none; border-radius: 8px; font-weight: 600; cursor: pointer; margin-bottom: 8px;';
    confirmBtn.addEventListener('click', () => {
      if (!consentCheckbox.checked) {
        this._setState({ ...s, error: 'You must authorize the hard credit inquiry to continue.' });
        return;
      }
      this._confirm(true);
    });
    wrap.appendChild(confirmBtn);

    const declineBtn = document.createElement('button');
    declineBtn.textContent = 'Decline This Offer';
    declineBtn.disabled = !!s.submitting;
    declineBtn.style.cssText = 'width: 100%; padding: 10px; background: transparent; color: #6b7280; border: 1px solid #e5e7eb; border-radius: 8px; cursor: pointer;';
    declineBtn.addEventListener('click', () => this._confirm(false));
    wrap.appendChild(declineBtn);

    this.appendChild(wrap);
  }
}
