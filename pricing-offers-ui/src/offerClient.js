export async function fetchSelectedOffer(apiBaseUrl, applicationId, sessionToken) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/selected-offer`, {
    headers: { Authorization: `Bearer ${sessionToken}` },
  });
  if (!res.ok) {
    const err = new Error(`Failed to load offer (HTTP ${res.status})`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}

export async function confirmSelectedOffer(apiBaseUrl, applicationId, sessionToken, selectedOfferId, consentGiven) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/selected-offer/confirm`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${sessionToken}` },
    body: JSON.stringify({ selectedOfferId, consentGiven }),
  });
  if (!res.ok) {
    throw new Error(`Failed to confirm offer (HTTP ${res.status})`);
  }
  return res.json();
}
