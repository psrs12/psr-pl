import { DocumentUploadManager } from './DocumentUploadManager.js';

if (!customElements.get('document-upload-manager')) {
  customElements.define('document-upload-manager', DocumentUploadManager);
}
