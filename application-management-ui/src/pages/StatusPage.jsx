import { useEffect, useRef, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import Header from '../components/Header.jsx';
import { getApplication } from '../api/client.js';
import { useWebComponentScript } from '../hooks/useWebComponent.js';
import { API } from '../api/config.js';
import { PROCESSING_STATES, UNDER_REVIEW_STATES } from '../navigation/navigationConfig.js';
import { resolveScreen, UnmappedApplicationStatusError } from '../navigation/resolveScreen.js';
import { SCREEN_REGISTRY } from '../navigation/screenRegistry.js';

const POLL_INTERVAL_MS = 8000;

export default function StatusPage() {
  const navigate = useNavigate();
  const applicationId = sessionStorage.getItem('applicationId');

  const offerConfirmedKey = `offerConfirmed:${applicationId}`;

  const [application, setApplication] = useState(null);
  const [loadError, setLoadError] = useState(null);
  const pollRef = useRef(null);

  const pricingScript = useWebComponentScript(API.pricingOffersUiJs);
  const documentScript = useWebComponentScript(API.documentManagementUiJs);
  const offerAcceptanceScript = useWebComponentScript(API.offerAcceptanceUiJs);

  useEffect(() => {
    if (!applicationId) {
      navigate('/portal/login');
      return;
    }
    fetchStatus();
    return () => { if (pollRef.current) clearTimeout(pollRef.current); };
  }, []);

  async function fetchStatus() {
    try {
      const data = await getApplication(applicationId);
      const status = data.applicationStatus ?? data.status;
      setApplication(data);
      if (status !== 'OFFER_PENDING') {
        sessionStorage.removeItem(offerConfirmedKey);
      }
      if (shouldPoll(status)) {
        pollRef.current = setTimeout(fetchStatus, POLL_INTERVAL_MS);
      }
    } catch (e) {
      if (e.status === 401 || e.status === 403) {
        sessionStorage.clear();
        navigate('/portal/login');
        return;
      }
      setLoadError(e.message);
      pollRef.current = setTimeout(fetchStatus, POLL_INTERVAL_MS);
    }
  }

  function shouldPoll(status) {
    return (
      PROCESSING_STATES.has(status) ||
      UNDER_REVIEW_STATES.has(status) ||
      (status === 'OFFER_PENDING' && sessionStorage.getItem(offerConfirmedKey) === 'true')
    );
  }

  const handleOfferConfirmed = useCallback(() => {
    sessionStorage.setItem(offerConfirmedKey, 'true');
    if (pollRef.current) clearTimeout(pollRef.current);
    setTimeout(fetchStatus, 2000);
  }, []);

  const handleOfferAccepted = useCallback(() => {
    if (pollRef.current) clearTimeout(pollRef.current);
    setTimeout(fetchStatus, 2000);
  }, []);

  if (loadError && !application) {
    return (
      <div className="page-shell">
        <Header />
        <div className="card">
          <div className="alert-error">{loadError}</div>
          <button className="btn-primary" onClick={fetchStatus} style={{ marginTop: 16 }}>Retry</button>
        </div>
      </div>
    );
  }

  if (!application) {
    return (
      <div className="page-shell">
        <Header />
        <div className="status-center">
          <div className="spinner" />
          <p style={{ color: '#6b7280' }}>Loading your application...</p>
        </div>
      </div>
    );
  }

  const status = application.applicationStatus ?? application.status;

  let descriptor;
  let resolveError = null;
  try {
    descriptor = resolveScreen(status);
  } catch (e) {
    if (e instanceof UnmappedApplicationStatusError) {
      resolveError = e;
    } else {
      throw e;
    }
  }

  const ctx = {
    status,
    application,
    applicationId,
    scripts: { pricingOffersUiJs: pricingScript, documentManagementUiJs: documentScript, offerAcceptanceUiJs: offerAcceptanceScript },
    eventHandlers: { 'offer-confirmed': handleOfferConfirmed, 'offer-accepted': handleOfferAccepted },
  };

  const Screen = descriptor ? SCREEN_REGISTRY[descriptor.kind] : null;

  return (
    <div className="page-shell">
      <Header />
      {resolveError || !Screen
        ? renderUnmappedStatusFallback(resolveError, status)
        : <Screen descriptor={descriptor} ctx={ctx} />}
    </div>
  );
}

function renderUnmappedStatusFallback(error, status) {
  // eslint-disable-next-line no-console
  console.error(error ?? new Error(`No screen registered for status: ${status}`));
  return (
    <div className="status-center">
      <div className="spinner" />
      <h3 style={{ color: '#1a1a1a' }}>Processing</h3>
      <p style={{ color: '#6b7280' }}>Status: {status}</p>
    </div>
  );
}
