export async function fetchRequiredDocuments(apiBaseUrl, applicationId, sessionToken) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/required-documents`, {
    headers: { Authorization: `Bearer ${sessionToken}` },
  });
  if (!res.ok) {
    const err = new Error(`Failed to load required documents (HTTP ${res.status})`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}

export async function fetchDocumentRecord(apiBaseUrl, applicationId, sessionToken) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/documents`, {
    headers: { Authorization: `Bearer ${sessionToken}` },
  });
  if (!res.ok) {
    const err = new Error(`Failed to load document status (HTTP ${res.status})`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}

export async function submitDocument(apiBaseUrl, applicationId, sessionToken, documentId) {
  const res = await fetch(`${apiBaseUrl}/applications/${applicationId}/documents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${sessionToken}` },
    body: JSON.stringify({ documentId }),
  });
  if (!res.ok) {
    let detail = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      detail = body.message ?? detail;
    } catch (_) {}
    const err = new Error(`Failed to submit document: ${detail}`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}
