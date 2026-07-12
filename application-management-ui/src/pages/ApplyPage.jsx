import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import Header from '../components/Header.jsx';
import { validateInvitation, createApplication, login } from '../api/client.js';

const LOAN_PURPOSES = [
  { value: 'DEBT_CONSOLIDATION', label: 'Debt Consolidation' },
  { value: 'HOME_IMPROVEMENT', label: 'Home Improvement' },
  { value: 'MAJOR_PURCHASE', label: 'Major Purchase' },
  { value: 'MEDICAL', label: 'Medical' },
  { value: 'VEHICLE', label: 'Vehicle' },
  { value: 'BUSINESS', label: 'Business' },
  { value: 'OTHER', label: 'Other' },
];

const EMPLOYMENT_TYPES = [
  { value: 'FULL_TIME', label: 'Full Time' },
  { value: 'PART_TIME', label: 'Part Time' },
  { value: 'SELF_EMPLOYED', label: 'Self Employed' },
  { value: 'CONTRACT', label: 'Contract' },
  { value: 'RETIRED', label: 'Retired' },
  { value: 'UNEMPLOYED', label: 'Unemployed' },
];

const US_STATES = [
  'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA',
  'HI','ID','IL','IN','IA','KS','KY','LA','ME','MD',
  'MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ',
  'NM','NY','NC','ND','OH','OK','OR','PA','RI','SC',
  'SD','TN','TX','UT','VT','VA','WA','WV','WI','WY',
];

const TERM_OPTIONS = [36, 48, 60, 72, 84];

const CITIZENSHIP_STATUSES = [
  { value: 'US_CITIZEN', label: 'U.S. Citizen' },
  { value: 'PERMANENT_RESIDENT', label: 'Permanent Resident' },
  { value: 'VISA_HOLDER', label: 'Visa Holder' },
  { value: 'OTHER', label: 'Other' },
];

const EMPTY_FORM = {
  firstName: '',
  lastName: '',
  dateOfBirth: '',
  ssn: '',
  citizenship: 'US_CITIZEN',
  phone: '',
  email: '',
  addressLine1: '',
  addressLine2: '',
  city: '',
  state: '',
  postcode: '',
  employmentType: 'FULL_TIME',
  employerName: '',
  annualIncome: '',
  monthlyObligations: '',
  requestedAmount: '',
  requestedTermMonths: 36,
  loanPurpose: 'DEBT_CONSOLIDATION',
};

/* ─── Step bar ─────────────────────────────────────────────── */
function Steps({ current }) {
  const steps = ['Personal Info', 'Employment Info', 'Get Qualified'];
  return (
    <div className="step-bar">
      <div className="step-list">
        {steps.map((label, i) => {
          const num = i + 1;
          const active = num === current;
          const done = num < current;
          return (
            <React.Fragment key={label}>
              <div className="step-item">
                <div className={`step-bubble${active ? ' step-bubble--active' : done ? ' step-bubble--done' : ''}`}>
                  {done ? '✓' : num}
                </div>
                <span className={`step-label${active ? ' step-label--active' : ''}`}>{label}</span>
              </div>
              {i < steps.length - 1 && <div className="step-connector" />}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
}

/* ─── Invitation dropdown ───────────────────────────────────── */
function InvitationPanel({ open, onPrefillSuccess }) {
  const [token, setToken] = useState('');
  const [validating, setValidating] = useState(false);
  const [error, setError] = useState(null);
  const inputRef = useRef(null);

  React.useEffect(() => {
    if (open) setTimeout(() => inputRef.current?.focus(), 50);
  }, [open]);

  async function handleApply() {
    if (!token.trim()) return;
    setValidating(true);
    setError(null);
    try {
      const data = await validateInvitation(token.trim());
      onPrefillSuccess(data, token.trim());
      setToken('');
    } catch (e) {
      setError(e.message);
    } finally {
      setValidating(false);
    }
  }

  if (!open) return null;

  return (
    <div className="invitation-panel-wrapper">
      <div className="invitation-reveal">
        <input
          ref={inputRef}
          type="text"
          className="invitation-code-input"
          value={token}
          onChange={e => { setToken(e.target.value); setError(null); }}
          onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); handleApply(); } }}
          placeholder="_ _ _ _ _ _ _ _ _ _ _ _   _ _ _ _ _ _"
          disabled={validating}
          spellCheck={false}
        />
        {error && <div className="alert-error" style={{ marginTop: 0, textAlign: 'left', width: '100%', maxWidth: 520 }}>{error}</div>}
        <button
          type="button"
          className="btn-apply-offer"
          onClick={handleApply}
          disabled={validating || !token.trim()}
        >
          {validating ? 'Checking…' : 'Apply Offer'}
        </button>
      </div>
    </div>
  );
}

/* ─── Review row helper ─────────────────────────────────────── */
function ReviewRow({ label, value }) {
  if (!value) return null;
  return (
    <div className="review-row">
      <span className="review-label">{label}</span>
      <span className="review-value">{value}</span>
    </div>
  );
}

/* ─── Main page ─────────────────────────────────────────────── */
export default function ApplyPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [step, setStep] = useState(1);
  const [form, setForm] = useState(EMPTY_FORM);
  const [prefill, setPrefill] = useState(null);
  const [invitationToken, setInvitationToken] = useState('');
  const [invitationOpen, setInvitationOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState(null);
  const [submittedApplicationId, setSubmittedApplicationId] = useState(null);

  function set(field, value) {
    setForm(f => ({ ...f, [field]: value }));
  }

  const isITA = !!prefill;

  useEffect(() => {
    const urlToken = searchParams.get('token');
    if (urlToken && !prefill) {
      validateInvitation(urlToken)
        .then(data => handlePrefillSuccess(data, urlToken))
        .catch(() => {});
    }
  }, []);

  function handlePrefillSuccess(data, token) {
    setPrefill(data);
    setInvitationToken(token);
    setInvitationOpen(false);
    setForm(f => ({
      ...f,
      firstName: data.firstName ?? '',
      lastName: data.lastName ?? '',
      addressLine1: data.address?.line1 ?? '',
      city: data.address?.city ?? '',
      state: data.address?.state ?? '',
      postcode: data.address?.postcode ?? '',
    }));
  }

  async function handleSubmit() {
    setSubmitting(true);
    setSubmitError(null);
    try {
      const address = {
        line1: form.addressLine1,
        line2: form.addressLine2 || null,
        city: form.city,
        state: form.state,
        postcode: form.postcode,
      };
      const employment = {
        employmentType: form.employmentType,
        employerName: form.employerName,
        annualIncome: Number(form.annualIncome),
        monthlyObligations: Number(form.monthlyObligations),
      };
      const contact = { phone: form.phone, email: form.email };
      const payload = isITA
        ? { applicationSource: 'INVITATION', intakeId: prefill.intakeId, firstName: form.firstName, lastName: form.lastName, dateOfBirth: form.dateOfBirth, ssn: form.ssn, citizenship: form.citizenship, address, ...contact, requestedAmount: Number(form.requestedAmount), requestedTermMonths: Number(form.requestedTermMonths), loanPurpose: form.loanPurpose, ...employment }
        : { applicationSource: 'DIRECT', firstName: form.firstName, lastName: form.lastName, dateOfBirth: form.dateOfBirth, ssn: form.ssn, citizenship: form.citizenship, address, ...contact, requestedAmount: Number(form.requestedAmount), requestedTermMonths: Number(form.requestedTermMonths), loanPurpose: form.loanPurpose, ...employment };
      const response = await createApplication(payload);
      const last4SSN = form.ssn.replace(/\D/g, '').slice(-4);
      const session = await login(response.applicationId, last4SSN, form.dateOfBirth);
      sessionStorage.setItem('sessionToken', session.sessionToken);
      sessionStorage.setItem('applicationId', session.applicationId);
      setSubmittedApplicationId(response.applicationId);
      setStep(4);
      window.scrollTo(0, 0);
    } catch (e) {
      setSubmitError(e.message);
    } finally {
      setSubmitting(false);
    }
  }

  const purposeLabel = LOAN_PURPOSES.find(p => p.value === form.loanPurpose)?.label ?? '';
  const employmentLabel = EMPLOYMENT_TYPES.find(t => t.value === form.employmentType)?.label ?? '';

  return (
    <div className="page-shell">
      <Header onInvitationClick={() => setInvitationOpen(o => !o)} />
      <InvitationPanel open={invitationOpen} onPrefillSuccess={handlePrefillSuccess} />
      {step <= 3 && <Steps current={step} />}

      <div className="apply-page">

        {/* ── Hero (step 1 only) ── */}
        {step === 1 && (
          <>
            <div className="apply-hero">
              <h1>See if you <em>qualify</em> in minutes</h1>
              <p>No commitment — check your rate with <strong>no impact to your credit score</strong></p>
            </div>
            <div className="trust-strip">
              <div className="trust-badge">
                <div className="trust-badge-icon">✓</div>
                No credit score impact
              </div>
              <div className="trust-badge">
                <div className="trust-badge-icon">⚡</div>
                Decision in 2 minutes
              </div>
              <div className="trust-badge">
                <div className="trust-badge-icon">🔒</div>
                Bank-level security
              </div>
            </div>
          </>
        )}

        {isITA && step === 1 && (
          <div className="prefill-notice" style={{ marginBottom: 16 }}>
            Invitation validated — your name and address have been pre-filled and are read-only.
          </div>
        )}

        {/* ══════════════ STEP 1 — Personal Info ══════════════ */}
        {step === 1 && (
          <form onSubmit={e => { e.preventDefault(); setStep(2); window.scrollTo(0,0); }}>

            <div className="apply-section">
              <div className="apply-section-title">Let's start with some information about your loan</div>

              <div className="form-row">
                <div className="form-group">
                  <label>How much do you need?</label>
                  <input type="number" min="1000" step="100" value={form.requestedAmount}
                    onChange={e => set('requestedAmount', e.target.value)}
                    required placeholder="e.g. $15,000" />
                </div>
                <div className="form-group">
                  <label>What's the purpose?</label>
                  <select value={form.loanPurpose} onChange={e => set('loanPurpose', e.target.value)} required>
                    {LOAN_PURPOSES.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
                  </select>
                </div>
              </div>

              <div>
                <div className="term-label">What's the loan term? (months)</div>
                <div className="term-pills">
                  {TERM_OPTIONS.map(t => (
                    <button key={t} type="button"
                      className={`term-pill${form.requestedTermMonths === t ? ' term-pill--selected' : ''}`}
                      onClick={() => set('requestedTermMonths', t)}>
                      {t}
                    </button>
                  ))}
                </div>
                <div style={{ marginTop: 14, fontSize: '0.87rem', color: '#e8520a', fontWeight: 600 }}>
                  ⊕ Estimate Your Monthly Payment
                </div>
              </div>
            </div>

            <div className="apply-section">
              <div className="apply-section-title">Next, a few details about you</div>

              <div className="form-row">
                <div className="form-group">
                  <label>First Name</label>
                  <input type="text" value={form.firstName} onChange={e => set('firstName', e.target.value)}
                    readOnly={isITA} required placeholder="First Name" />
                </div>
                <div className="form-group">
                  <label>Last Name</label>
                  <input type="text" value={form.lastName} onChange={e => set('lastName', e.target.value)}
                    readOnly={isITA} required placeholder="Last Name" />
                </div>
              </div>

              <div className="form-group">
                <label>Street Address</label>
                <input type="text" value={form.addressLine1} onChange={e => set('addressLine1', e.target.value)}
                  readOnly={isITA} required placeholder="Street Address — 47 W 5th St" />
              </div>

              <div className="form-group">
                <label>Apartment / Suite <span style={{ fontWeight: 400, color: '#9ca3af' }}>(optional)</span></label>
                <input type="text" value={form.addressLine2} onChange={e => set('addressLine2', e.target.value)}
                  placeholder="Apt, Suite, Unit, etc." />
              </div>

              <div className="form-row-3">
                <div className="form-group">
                  <label>City</label>
                  <input type="text" value={form.city} onChange={e => set('city', e.target.value)}
                    readOnly={isITA} required placeholder="City" />
                </div>
                <div className="form-group">
                  <label>State</label>
                  <select value={form.state} onChange={e => set('state', e.target.value)}
                    disabled={isITA} required>
                    <option value="">State</option>
                    {US_STATES.map(s => <option key={s} value={s}>{s}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>ZIP Code</label>
                  <input type="text" value={form.postcode} onChange={e => set('postcode', e.target.value)}
                    readOnly={isITA} required placeholder="ZIP Code" maxLength={10} />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Preferred Phone Number</label>
                  <input type="tel" value={form.phone} onChange={e => set('phone', e.target.value)}
                    required placeholder="(555) 000-0000" />
                </div>
                <div className="form-group">
                  <label>Email Address</label>
                  <input type="email" value={form.email} onChange={e => set('email', e.target.value)}
                    required placeholder="jane.smith@email.com" />
                </div>
              </div>
            </div>

            <button type="submit" className="btn-primary">Continue</button>
            <p className="btn-no-impact">This won't impact your credit score</p>
          </form>
        )}

        {/* ══════════════ STEP 2 — Employment Info ════════════ */}
        {step === 2 && (
          <form onSubmit={e => { e.preventDefault(); setStep(3); window.scrollTo(0,0); }}>

            <div className="apply-section">
              <div className="apply-section-title">Tell us about your employment &amp; income</div>

              <div className="form-row">
                <div className="form-group">
                  <label>Date of Birth</label>
                  <input type="date" value={form.dateOfBirth}
                    onChange={e => set('dateOfBirth', e.target.value)} required />
                </div>
                <div className="form-group">
                  <label>Social Security Number</label>
                  <input type="password" value={form.ssn} onChange={e => set('ssn', e.target.value)}
                    required placeholder="XXX-XX-XXXX" autoComplete="off" />
                </div>
              </div>

              <div className="form-group">
                <label>Citizenship Status</label>
                <select value={form.citizenship} onChange={e => set('citizenship', e.target.value)} required>
                  {CITIZENSHIP_STATUSES.map(c => <option key={c.value} value={c.value}>{c.label}</option>)}
                </select>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Employment Type</label>
                  <select value={form.employmentType} onChange={e => set('employmentType', e.target.value)} required>
                    {EMPLOYMENT_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>Employer Name <span style={{ fontWeight: 400, color: '#9ca3af' }}>(optional)</span></label>
                  <input type="text" value={form.employerName} onChange={e => set('employerName', e.target.value)}
                    placeholder="e.g. Acme Corp" />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Annual Income ($)</label>
                  <input type="number" min="0" step="1" value={form.annualIncome}
                    onChange={e => set('annualIncome', e.target.value)} required placeholder="e.g. 65,000" />
                </div>
                <div className="form-group">
                  <label>Monthly Obligations ($)</label>
                  <input type="number" min="0" step="1" value={form.monthlyObligations}
                    onChange={e => set('monthlyObligations', e.target.value)} required placeholder="e.g. 800" />
                </div>
              </div>
            </div>

            <div className="step-nav">
              <button type="button" className="btn-back" onClick={() => { setStep(1); window.scrollTo(0,0); }}>
                ← Back
              </button>
              <button type="submit" className="btn-primary" style={{ flex: 1 }}>Continue</button>
            </div>
            <p className="btn-no-impact">This won't impact your credit score</p>
          </form>
        )}

        {/* ══════════════ STEP 3 — Get Qualified ══════════════ */}
        {step === 3 && (
          <div>
            <div className="apply-section">
              <div className="apply-section-title">Review your application</div>
              <p style={{ color: '#6b7280', fontSize: '0.92rem', textAlign: 'center', marginBottom: 24 }}>
                Please confirm your details before submitting.
              </p>

              <div className="review-group-label">Loan Details</div>
              <div className="review-block">
                <ReviewRow label="Loan Amount" value={form.requestedAmount ? `$${Number(form.requestedAmount).toLocaleString()}` : ''} />
                <ReviewRow label="Purpose" value={purposeLabel} />
                <ReviewRow label="Term" value={form.requestedTermMonths ? `${form.requestedTermMonths} months` : ''} />
              </div>

              <div className="review-group-label">Personal Details</div>
              <div className="review-block">
                <ReviewRow label="Name" value={[form.firstName, form.lastName].filter(Boolean).join(' ')} />
                <ReviewRow label="Address" value={[form.addressLine1, form.addressLine2, form.city, form.state, form.postcode].filter(Boolean).join(', ')} />
                <ReviewRow label="Phone" value={form.phone} />
                <ReviewRow label="Email" value={form.email} />
                <ReviewRow label="Citizenship" value={CITIZENSHIP_STATUSES.find(c => c.value === form.citizenship)?.label} />
              </div>

              <div className="review-group-label">Employment &amp; Income</div>
              <div className="review-block">
                <ReviewRow label="Employment" value={employmentLabel} />
                {form.employerName && <ReviewRow label="Employer" value={form.employerName} />}
                <ReviewRow label="Annual Income" value={form.annualIncome ? `$${Number(form.annualIncome).toLocaleString()}` : ''} />
                <ReviewRow label="Monthly Obligations" value={form.monthlyObligations ? `$${Number(form.monthlyObligations).toLocaleString()}` : ''} />
              </div>
            </div>

            {submitError && <div className="alert-error">{submitError}</div>}

            <div className="step-nav">
              <button type="button" className="btn-back" onClick={() => { setStep(2); window.scrollTo(0,0); }}>
                ← Back
              </button>
              <button type="button" className="btn-primary" style={{ flex: 1 }}
                disabled={submitting} onClick={handleSubmit}>
                {submitting ? 'Submitting…' : 'Submit Application'}
              </button>
            </div>
            <p className="btn-no-impact">This won't impact your credit score</p>
          </div>
        )}

        {/* ══════════════ STEP 4 — Confirmation ══════════════ */}
        {step === 4 && (
          <div className="apply-section" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '2.8rem', marginBottom: 10 }}>✓</div>
            <h2 style={{ marginBottom: 8 }}>Application Submitted</h2>
            <p style={{ color: '#6b7280', marginBottom: 20 }}>
              We're already checking your rates with no impact to your credit score.
              Save your Application ID below — you'll need it to check your status.
            </p>
            <div style={{
              display: 'inline-block',
              background: '#fff8f5',
              border: '1.5px dashed #fbd5c0',
              borderRadius: 10,
              padding: '14px 28px',
              fontSize: '1.15rem',
              fontWeight: 700,
              letterSpacing: '0.02em',
              color: '#e8520a',
              marginBottom: 24,
            }}>
              {submittedApplicationId}
            </div>
            <div>
              <button
                type="button"
                className="btn-primary"
                style={{ width: 'auto', padding: '15px 40px' }}
                onClick={() => navigate('/portal/status')}
              >
                View My Offers
              </button>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
