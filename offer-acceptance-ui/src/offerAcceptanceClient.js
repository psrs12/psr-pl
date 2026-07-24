export async function fetchConfirmedOffer(pricingApiBaseUrl, applicationId, sessionToken) {
  const res = await fetch(`${pricingApiBaseUrl}/applications/${applicationId}/selected-offer`, {
    headers: { Authorization: `Bearer ${sessionToken}` },
  });
  if (!res.ok) {
    const err = new Error(`Failed to load confirmed offer (HTTP ${res.status})`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}

export async function fetchDeclarations(apiBaseUrl, applicationId, sessionToken) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/declarations`, {
    headers: { Authorization: `Bearer ${sessionToken}` },
  });
  if (!res.ok) {
    const err = new Error(`Failed to load declarations (HTTP ${res.status})`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}

export async function submitEsign(apiBaseUrl, applicationId, sessionToken, acceptedDeclarationIds) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/esign`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${sessionToken}` },
    body: JSON.stringify({ acceptedDeclarationIds }),
  });
  if (!res.ok) {
    let detail = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      detail = body.message ?? detail;
    } catch (_) {}
    const err = new Error(`Failed to e-sign: ${detail}`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}
