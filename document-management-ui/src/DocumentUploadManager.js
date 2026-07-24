import { fetchRequiredDocuments, fetchDocumentRecord, submitDocument } from './documentClient.js';

// <document-upload-manager application-id="..." api-base-url="..." session-token="...">
// Shown once the application reaches DOCUMENTS_REQUIRED (right after
// e-sign). There is no real file upload/storage here (document-service
// has no S3 integration) -- "Mark as Submitted" just records that a given
// document id was provided, which is enough to exercise the
// DOCUMENTS_REQUIRED -> OFFER_ACCEPTED transition end to end. Once every
// required document is submitted, dispatches 'documents-submitted' so the
// host shell (application-management-ui's StatusPage) knows to refresh.
export class DocumentUploadManager extends HTMLElement {
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
      const [required, record] = await Promise.all([
        fetchRequiredDocuments(this.apiBaseUrl, this.applicationId, this.sessionToken),
        fetchDocumentRecord(this.apiBaseUrl, this.applicationId, this.sessionToken),
      ]);
      this._setState({ loading: false, required, record });
    } catch (err) {
      this._setState({ loading: false, error: err.message });
    }
  }

  async _submit(documentId) {
    const s = this._state;
    this._setState({ ...s, submittingId: documentId, error: null });
    try {
      const record = await submitDocument(this.apiBaseUrl, this.applicationId, this.sessionToken, documentId);
      this._setState({ ...s, submittingId: null, record });
      if (record.status === 'COMPLETE') {
        this.dispatchEvent(new CustomEvent('documents-submitted', { bubbles: true }));
      }
    } catch (err) {
      this._setState({ ...s, submittingId: null, error: err.message });
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
    wrap.className = 'document-upload-manager';
    wrap.style.cssText = 'font-family: system-ui, sans-serif; max-width: 640px; margin: 0 auto; padding: 24px; border: 1px solid #e5e7eb; border-radius: 12px;';

    if (s.error) {
      const err = document.createElement('div');
      err.style.cssText = 'color: #b71c1c; margin-bottom: 12px;';
      err.textContent = s.error;
      wrap.appendChild(err);
    }

    if (s.loading) {
      const loading = document.createElement('p');
      loading.textContent = 'Loading required documents…';
      wrap.appendChild(loading);
      this.appendChild(wrap);
      return;
    }

    if (!s.required || !s.record) {
      this.appendChild(wrap);
      return;
    }

    const title = document.createElement('h2');
    title.textContent = 'Documents Needed';
    title.style.marginTop = '0';
    wrap.appendChild(title);

    const intro = document.createElement('p');
    intro.style.cssText = 'color: #4b5563; margin-bottom: 20px;';
    intro.textContent = 'We need a few documents to finish processing your loan.';
    wrap.appendChild(intro);

    const submittedIds = new Set(s.record.submittedDocumentIds || []);

    s.required.forEach((doc) => {
      const submitted = submittedIds.has(doc.id);

      const item = document.createElement('div');
      item.style.cssText = 'display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; border: 1px solid #e5e7eb; border-radius: 8px; padding: 14px 16px; margin-bottom: 10px;';

      const textWrap = document.createElement('div');
      const heading = document.createElement('strong');
      heading.textContent = doc.title + (doc.required ? ' (required)' : ' (optional)');
      textWrap.appendChild(heading);
      const body = document.createElement('p');
      body.style.cssText = 'margin: 4px 0 0; font-size: 0.88rem; color: #4b5563;';
      body.textContent = doc.description;
      textWrap.appendChild(body);
      item.appendChild(textWrap);

      if (submitted) {
        const badge = document.createElement('span');
        badge.textContent = '✓ Submitted';
        badge.style.cssText = 'color: #15803d; font-weight: 600; white-space: nowrap; align-self: center;';
        item.appendChild(badge);
      } else {
        const btn = document.createElement('button');
        btn.textContent = s.submittingId === doc.id ? 'Submitting…' : 'Mark as Submitted';
        btn.disabled = s.submittingId === doc.id;
        btn.style.cssText = 'padding: 8px 14px; background: #2563eb; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; white-space: nowrap; align-self: center;';
        btn.addEventListener('click', () => this._submit(doc.id));
        item.appendChild(btn);
      }

      wrap.appendChild(item);
    });

    if (s.record.status === 'COMPLETE') {
      const done = document.createElement('p');
      done.style.cssText = 'color: #15803d; font-weight: 600; margin-top: 16px;';
      done.textContent = 'All documents received — moving on to final verification.';
      wrap.appendChild(done);
    }

    this.appendChild(wrap);
  }
}
