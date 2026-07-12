import { API } from './config.js';

function getSession() {
  return {
    token: sessionStorage.getItem('sessionToken'),
    applicationId: sessionStorage.getItem('applicationId'),
  };
}

function authHeaders() {
  const { token } = getSession();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse(res) {
  if (!res.ok) {
    let detail = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      detail = body.detail ?? body.message ?? detail;
    } catch (_) {}
    const err = new Error(detail);
    err.status = res.status;
    throw err;
  }
  if (res.status === 204) return null;
  return res.json();
}

export async function validateInvitation(token) {
  const res = await fetch(`${API.appManagement}/invitations/validate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  return handleResponse(res);
}

export async function createApplication(payload) {
  const isITA = payload.applicationSource === 'INVITATION';
  const channelId = isITA ? 'ITA' : 'WEB';

  const body = {
    intakeId: payload.intakeId ?? null,
    ssnVerificationToken: payload.ssnVerificationToken ?? null,
    ssn: payload.ssn ? payload.ssn.replace(/\D/g, '') : null,
    firstName: payload.firstName,
    lastName: payload.lastName,
    dateOfBirth: payload.dateOfBirth ?? null,
    citizenship: payload.citizenship,
    email: payload.email,
    phone: payload.phone,
    street: payload.address?.line1 ?? payload.street,
    addressLine2: payload.address?.line2 ?? payload.addressLine2 ?? null,
    city: payload.address?.city ?? payload.city,
    state: payload.address?.state ?? payload.state,
    zip: payload.address?.postcode ?? payload.zip,
    employerName: payload.employerName ?? null,
    employmentStatus: mapEmploymentStatus(payload.employmentType ?? payload.employmentStatus),
    annualIncome: payload.annualIncome,
    monthlyObligations: payload.monthlyObligations,
    requestedAmount: payload.requestedAmount,
    termMonths: payload.requestedTermMonths ?? payload.termMonths,
    loanPurpose: payload.loanPurpose,
  };

  const res = await fetch(`${API.appManagement}/applications`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Channel-ID': channelId },
    body: JSON.stringify(body),
  });
  return handleResponse(res);
}

function mapEmploymentStatus(type) {
  switch (type) {
    case 'FULL_TIME':
    case 'PART_TIME':
    case 'CONTRACT': return 'EMPLOYED';
    case 'SELF_EMPLOYED': return 'SELF_EMPLOYED';
    case 'RETIRED': return 'RETIRED';
    default: return 'OTHER';
  }
}

export async function login(applicationId, last4SSN, dateOfBirth) {
  const res = await fetch(`${API.appManagement}/applications/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ applicationId, last4SSN, dateOfBirth }),
  });
  return handleResponse(res);
}

export async function getApplication(applicationId) {
  const res = await fetch(`${API.appManagement}/applications/${applicationId}/status`, {
    headers: authHeaders(),
    cache: 'no-store',
  });
  return handleResponse(res);
}

export async function getSelectedOffer(applicationId) {
  const res = await fetch(`${API.pricing}/applications/${applicationId}/selected-offer`, {
    headers: authHeaders(),
    cache: 'no-store',
  });
  return handleResponse(res);
}

export async function getDeclarations(applicationId) {
  const res = await fetch(`${API.offerAcceptance}/applications/${applicationId}/declarations`, {
    headers: authHeaders(),
    cache: 'no-store',
  });
  return handleResponse(res);
}

export async function submitESign(applicationId, acceptedDeclarationIds) {
  const res = await fetch(`${API.offerAcceptance}/applications/${applicationId}/esign`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ acceptedDeclarationIds }),
  });
  return handleResponse(res);
}
