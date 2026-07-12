import { useRef } from 'react';
import { useCustomEvent } from '../../hooks/useWebComponent.js';
import { API } from '../../api/config.js';

export default function WebComponentScreen({ descriptor, ctx }) {
  const ref = useRef(null);
  const handler = descriptor.completionEvent ? ctx.eventHandlers?.[descriptor.completionEvent] : undefined;
  useCustomEvent(ref, descriptor.completionEvent ?? '__none__', handler ?? (() => {}));

  const script = ctx.scripts[descriptor.scriptUrlKey];

  if (script.error) {
    return (
      <div className="card">
        <div className="alert-error">Unable to load {descriptor.tag}: {script.error}</div>
      </div>
    );
  }

  if (!script.loaded) {
    return (
      <div className="status-center">
        <div className="spinner" />
        <p style={{ color: '#6b7280' }}>Loading...</p>
      </div>
    );
  }

  const attributeValues = {
    'application-id': ctx.applicationId,
    'api-base-url': API[descriptor.apiBaseUrlKey],
    'session-token': sessionStorage.getItem('sessionToken'),
    'pricing-api-base-url': API.pricing,
  };
  const props = Object.fromEntries(descriptor.attributes.map((name) => [name, attributeValues[name]]));

  const Tag = descriptor.tag;
  return (
    <div style={{ maxWidth: descriptor.containerWidth, margin: '40px auto', padding: '0 20px' }}>
      <Tag ref={ref} {...props} />
    </div>
  );
}
