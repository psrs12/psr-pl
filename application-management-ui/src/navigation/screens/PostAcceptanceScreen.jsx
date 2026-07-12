export default function PostAcceptanceScreen({ descriptor }) {
  return (
    <div className="status-center">
      <div style={{ fontSize: '3.5rem', marginBottom: 16 }}>✓</div>
      <h2 style={{ color: '#15803d' }}>{descriptor.label}</h2>
      <p style={{ color: '#6b7280', maxWidth: 460, margin: '0 auto' }}>{descriptor.message}</p>
    </div>
  );
}
