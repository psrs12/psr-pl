import { fetchSelectedOffer, confirmSelectedOffer } from './offerClient.js';
import { centsToDollarString } from './formatting.js';

// <pricing-offer-selector application-id="..." api-base-url="..." session-token="...">
// Two-step flow: 'list' shows the priced offer(s) (pricing-orchestration-
// service currently prices exactly one, rendered as a one-item list —
// the UI is structured to support more without a rewrite if that changes)
// with a Select button per offer; selecting moves to 'consent', which
// requires explicit hard-pull consent before confirming (no
// default-checked consent box — this is an FCRA consent gate, not a
// formality) and dispatches 'offer-confirmed' so the host shell
// (application-management-ui's StatusPage) knows to refresh.
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

  // Retries on 404 for a few seconds: workflow-status-service reports
  // OFFER_PENDING as soon as Step Functions *enters* AwaitOfferSelection,
  // which can race ahead of present-offer-lambda actually finishing its
  // call to store the offer -- a fetch landing in that gap 404s even
  // though the offer is genuinely about to exist, not actually missing.
  async _load() {
    if (!this.applicationId || !this.apiBaseUrl || !this.sessionToken) return;
    this._setState({ step: 'list', loading: true, error: null });
    const maxAttempts = 6;
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        const selected = await fetchSelectedOffer(this.apiBaseUrl, this.applicationId, this.sessionToken);
        this._setState({ step: 'list', loading: false, offers: selected.offers || [] });
        return;
      } catch (err) {
        const isLastAttempt = attempt === maxAttempts;
        if (err.status === 404 && !isLastAttempt) {
          await new Promise((resolve) => setTimeout(resolve, 1000));
          continue;
        }
        this._setState({ step: 'list', loading: false, error: err.message });
        return;
      }
    }
  }

  _select(offer) {
    this._setState({ ...this._state, step: 'consent', selectedOffer: offer, error: null });
  }

  async _confirm(consentGiven) {
    const s = this._state;
    this._setState({ ...s, submitting: true, error: null });
    try {
      await confirmSelectedOffer(this.apiBaseUrl, this.applicationId, this.sessionToken, s.selectedOffer.offerId, consentGiven);
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
    const s = this._state || { step: 'list', loading: true };
    this.innerHTML = '';

    const isConsentStep = s.step === 'consent' && s.selectedOffer;
    const wrap = document.createElement('div');
    wrap.className = 'pricing-offer-selector';
    wrap.style.cssText = `font-family: system-ui, sans-serif; max-width: ${isConsentStep ? 480 : 640}px; margin: 0 auto; padding: 24px; border: 1px solid #e5e7eb; border-radius: 12px;`;

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

    if (isConsentStep) {
      this._renderConsent(wrap, s);
    } else {
      this._renderList(wrap, s);
    }

    this.appendChild(wrap);
  }

  _renderList(wrap, s) {
    const title = document.createElement('h2');
    title.textContent = 'Your Offers';
    title.style.marginTop = '0';
    wrap.appendChild(title);

    const offers = s.offers || [];
    if (offers.length === 0) {
      const none = document.createElement('p');
      none.textContent = 'No offer is available right now.';
      wrap.appendChild(none);
      return;
    }

    const table = document.createElement('table');
    table.style.cssText = 'width: 100%; border-collapse: collapse; font-size: 0.92rem;';

    const thead = document.createElement('thead');
    thead.innerHTML = `
      <tr style="border-bottom: 2px solid #e5e7eb; text-align: left;">
        <th style="padding: 8px 6px;">Offer</th>
        <th style="padding: 8px 6px;">Amount</th>
        <th style="padding: 8px 6px;">Term</th>
        <th style="padding: 8px 6px;">APR</th>
        <th style="padding: 8px 6px;"></th>
      </tr>`;
    table.appendChild(thead);

    const tbody = document.createElement('tbody');
    offers.forEach((offer) => {
      const row = document.createElement('tr');
      row.style.cssText = 'border-bottom: 1px solid #e5e7eb;';

      const labelCell = document.createElement('td');
      labelCell.style.cssText = 'padding: 10px 6px; font-weight: 600;';
      labelCell.textContent = offer.label || '';
      row.appendChild(labelCell);

      const amountCell = document.createElement('td');
      amountCell.style.cssText = 'padding: 10px 6px;';
      amountCell.textContent = centsToDollarString(offer.amountCents);
      row.appendChild(amountCell);

      const termCell = document.createElement('td');
      termCell.style.cssText = 'padding: 10px 6px;';
      termCell.textContent = `${offer.termMonths} mo`;
      row.appendChild(termCell);

      const aprCell = document.createElement('td');
      aprCell.style.cssText = 'padding: 10px 6px;';
      aprCell.textContent = `${offer.aprPercentage}%`;
      row.appendChild(aprCell);

      const actionCell = document.createElement('td');
      actionCell.style.cssText = 'padding: 8px 6px; text-align: right;';
      const selectBtn = document.createElement('button');
      selectBtn.textContent = 'Select';
      selectBtn.style.cssText = 'padding: 8px 16px; background: #2563eb; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; white-space: nowrap;';
      selectBtn.addEventListener('click', () => this._select(offer));
      actionCell.appendChild(selectBtn);
      row.appendChild(actionCell);

      tbody.appendChild(row);
    });
    table.appendChild(tbody);
    wrap.appendChild(table);
  }

  _renderConsent(wrap, s) {
    const offer = s.selectedOffer;

    const title = document.createElement('h2');
    title.textContent = 'Authorize Hard Credit Inquiry';
    title.style.marginTop = '0';
    wrap.appendChild(title);

    const summary = document.createElement('p');
    const offerName = offer.label ? `${offer.label} — ` : '';
    summary.innerHTML = `<strong>Selected offer:</strong> ${offerName}${centsToDollarString(offer.amountCents)} over ${offer.termMonths} months at ${offer.aprPercentage}% APR`;
    wrap.appendChild(summary);

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
    declineBtn.style.cssText = 'width: 100%; padding: 10px; background: transparent; color: #6b7280; border: 1px solid #e5e7eb; border-radius: 8px; cursor: pointer; margin-bottom: 8px;';
    declineBtn.addEventListener('click', () => this._confirm(false));
    wrap.appendChild(declineBtn);

    const backBtn = document.createElement('button');
    backBtn.textContent = '← Back to Offers';
    backBtn.disabled = !!s.submitting;
    backBtn.style.cssText = 'width: 100%; padding: 8px; background: transparent; color: #9ca3af; border: none; cursor: pointer; font-size: 0.85rem;';
    backBtn.addEventListener('click', () => this._setState({ ...s, step: 'list', error: null }));
    wrap.appendChild(backBtn);
  }
}
