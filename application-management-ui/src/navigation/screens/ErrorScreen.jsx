export default function ErrorScreen() {
  return (
    <div className="status-center" style={{ maxWidth: 520 }}>
      <div style={{ fontSize: '3rem', color: '#b71c1c', marginBottom: 16 }}>⚠</div>
      <h2 style={{ color: '#b71c1c' }}>Something Went Wrong</h2>
      <p style={{ color: '#6b7280' }}>
        We ran into a problem processing your application. This is on our end, not something you did.
      </p>
      <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
        Please contact our support team so we can help resolve this.
      </p>
    </div>
  );
}
