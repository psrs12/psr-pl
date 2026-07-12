import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import Header from '../components/Header.jsx';
import { login } from '../api/client.js';

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [form, setForm] = useState({
    applicationId: location.state?.applicationId ?? '',
    last4SSN: '',
    dateOfBirth: '',
  });
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  function set(field, value) {
    setForm(f => ({ ...f, [field]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const data = await login(form.applicationId.trim(), form.last4SSN.trim(), form.dateOfBirth);
      sessionStorage.setItem('sessionToken', data.sessionToken);
      sessionStorage.setItem('applicationId', data.applicationId);
      navigate('/portal/status');
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page-shell">
      <Header />
      <div className="card">
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <div style={{ width: 54, height: 54, borderRadius: '50%', background: '#fff5f0', border: '2px solid #fbd5c0', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '1.4rem', margin: '0 auto 14px' }}>🔒</div>
          <h2 style={{ fontSize: '1.45rem', fontWeight: 800, letterSpacing: '-0.02em', marginBottom: 6 }}>Welcome back</h2>
          <p style={{ color: '#6b7280', fontSize: '0.9rem', margin: 0 }}>Enter your details to securely check your application status.</p>
        </div>
        <hr style={{ border: 'none', borderTop: '1px solid #f0f1f3', margin: '0 0 22px' }} />

        {error && <div className="alert-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Application ID</label>
            <input
              type="text"
              value={form.applicationId}
              onChange={e => set('applicationId', e.target.value)}
              required
              placeholder="e.g. APP-12345"
              autoComplete="off"
            />
          </div>

          <div className="form-group">
            <label>Last 4 Digits of SSN</label>
            <input
              type="password"
              maxLength={4}
              value={form.last4SSN}
              onChange={e => set('last4SSN', e.target.value)}
              required
              placeholder="XXXX"
              autoComplete="off"
              inputMode="numeric"
            />
          </div>

          <div className="form-group">
            <label>Date of Birth</label>
            <input
              type="date"
              value={form.dateOfBirth}
              onChange={e => set('dateOfBirth', e.target.value)}
              required
            />
          </div>

          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Verifying...' : 'Access My Application'}
          </button>
        </form>

        <p style={{ marginTop: 20, fontSize: '0.82rem', color: '#9ca3af', textAlign: 'center', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
          <span>🔒</span> Your session is encrypted and expires after 30 minutes.
        </p>
      </div>
    </div>
  );
}
