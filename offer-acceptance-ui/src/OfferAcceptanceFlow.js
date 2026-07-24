import { fetchConfirmedOffer, fetchDeclarations, submitEsign } from './offerAcceptanceClient.js';
import { centsToDollarString } from './formatting.js';

// <offer-acceptance-flow application-id="..." api-base-url="..."
//   session-token="..." pricing-api-base-url="...">
// Shown once the application reaches APPROVED. Loads the confirmed offer
// (from pricing-orchestration-service, via pricing-api-base-url) so the
// applicant can see exactly what they're signing for, loads the fixed
// declaration set (from offer-acceptance-service, via api-base-url),
// requires every required declaration to be checked before enabling
// e-sign (no default-checked boxes -- explicit acceptance only, same
// principle as pricing-offer-selector's hard-pull consent gate), then
// dispatches 'offer-accepted' so the host shell (application-management-
// ui's StatusPage) knows to refresh.
export class OfferAcceptanceFlow extends HTMLElement {
  static get observedAttributes() {
    return ['application-id', 'api-base-url', 'session-token', 'pricing-api-base-url'];
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
  get pricingApiBaseUrl() { return this.getAttribute('pricing-api-base-url'); }

  async _load() {
    if (!this.applicationId || !this.apiBaseUrl || !this.sessionToken || !this.pricingApiBaseUrl) return;
    this._setState({ loading: true, error: null });
    try {
      const [selectedOfferRecord, declarations] = await Promise.all([
        fetchConfirmedOffer(this.pricingApiBaseUrl, this.applicationId, this.sessionToken),
        fetchDeclarations(this.apiBaseUrl, this.applicationId, this.sessionToken),
      ]);
      // selectedOfferRecord is {offers: [...], selectedOfferId, ...} -- pull
      // out the one the applicant actually confirmed, not the whole list.
      const confirmedOffer = (selectedOfferRecord.offers || []).find(
        (o) => o.offerId === selectedOfferRecord.selectedOfferId
      );
      this._setState({
        loading: false,
        confirmedOffer,
        declarations,
        accepted: new Set(),
      });
    } catch (err) {
      this._setState({ loading: false, error: err.message });
    }
  }

  _toggle(declarationId, checked) {
    const s = this._state;
    const accepted = new Set(s.accepted);
    if (checked) accepted.add(declarationId);
    else accepted.delete(declarationId);
    this._setState({ ...s, accepted, error: null });
  }

  async _submit() {
    const s = this._state;
    const missing = s.declarations.filter((d) => d.required && !s.accepted.has(d.id));
    if (missing.length > 0) {
      this._setState({ ...s, error: 'You must accept every required declaration to continue.' });
      return;
    }
    this._setState({ ...s, submitting: true, error: null });
    try {
      await submitEsign(this.apiBaseUrl, this.applicationId, this.sessionToken, Array.from(s.accepted));
      this.dispatchEvent(new CustomEvent('offer-accepted', { bubbles: true }));
    } catch (err) {
      this._setState({ ...s, submitting: false, error: err.message });
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
    wrap.className = 'offer-acceptance-flow';
    wrap.style.cssText = 'font-family: system-ui, sans-serif; max-width: 560px; margin: 0 auto; padding: 24px; border: 1px solid #e5e7eb; border-radius: 12px;';

    if (s.error) {
      const err = document.createElement('div');
      err.style.cssText = 'color: #b71c1c; margin-bottom: 12px;';
      err.textContent = s.error;
      wrap.appendChild(err);
    }

    if (s.loading) {
      const loading = document.createElement('p');
      loading.textContent = 'Loading your offer and disclosures…';
      wrap.appendChild(loading);
      this.appendChild(wrap);
      return;
    }

    if (!s.confirmedOffer || !s.declarations) {
      this.appendChild(wrap);
      return;
    }

    const title = document.createElement('h2');
    title.textContent = 'Accept Your Offer';
    title.style.marginTop = '0';
    wrap.appendChild(title);

    const offer = s.confirmedOffer;
    const summary = document.createElement('p');
    summary.innerHTML = `<strong>Amount:</strong> ${centsToDollarString(offer.amountCents)} &nbsp; ` +
      `<strong>Term:</strong> ${offer.termMonths} months &nbsp; ` +
      `<strong>APR:</strong> ${offer.aprPercentage}%`;
    summary.style.cssText = 'background: #f9fafb; border-radius: 8px; padding: 12px 16px; margin-bottom: 20px;';
    wrap.appendChild(summary);

    const declarationsTitle = document.createElement('h3');
    declarationsTitle.textContent = 'Review and Accept';
    wrap.appendChild(declarationsTitle);

    s.declarations.forEach((declaration) => {
      const item = document.createElement('div');
      item.style.cssText = 'border: 1px solid #e5e7eb; border-radius: 8px; padding: 12px 16px; margin-bottom: 10px;';

      const label = document.createElement('label');
      label.style.cssText = 'display: flex; align-items: flex-start; gap: 10px; cursor: pointer;';

      const checkbox = document.createElement('input');
      checkbox.type = 'checkbox';
      checkbox.checked = s.accepted.has(declaration.id);
      checkbox.style.marginTop = '3px';
      checkbox.addEventListener('change', (e) => this._toggle(declaration.id, e.target.checked));
      label.appendChild(checkbox);

      const textWrap = document.createElement('span');
      const heading = document.createElement('strong');
      heading.textContent = declaration.title + (declaration.required ? ' (required)' : ' (optional)');
      textWrap.appendChild(heading);
      const body = document.createElement('p');
      body.style.cssText = 'margin: 4px 0 0; font-size: 0.88rem; color: #4b5563;';
      body.textContent = declaration.text;
      textWrap.appendChild(body);
      label.appendChild(textWrap);

      item.appendChild(label);
      wrap.appendChild(item);
    });

    const submitBtn = document.createElement('button');
    submitBtn.textContent = s.submitting ? 'Submitting…' : 'I Accept and E-Sign';
    submitBtn.disabled = !!s.submitting;
    submitBtn.style.cssText = 'width: 100%; padding: 12px; background: #2563eb; color: white; border: none; border-radius: 8px; font-weight: 600; cursor: pointer; margin-top: 8px;';
    submitBtn.addEventListener('click', () => this._submit());
    wrap.appendChild(submitBtn);

    this.appendChild(wrap);
  }
}
