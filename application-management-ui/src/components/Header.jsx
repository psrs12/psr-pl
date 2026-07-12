import React from 'react';

export default function Header({ onInvitationClick }) {
  return (
    <header className="app-header">
      <span className="logo">Personal Loans</span>
      {onInvitationClick && (
        <button type="button" className="header-invitation-link header-invitation-center" onClick={onInvitationClick}>
          ✉ Have a Personal Invitation ID #?
        </button>
      )}
      <span className="header-right">Questions? Call us (1-800-555-0100)</span>
    </header>
  );
}
