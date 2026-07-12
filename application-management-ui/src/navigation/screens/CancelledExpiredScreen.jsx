export default function CancelledExpiredScreen({ descriptor }) {
  return (
    <div className="status-center">
      <h2 style={{ color: '#6b7280' }}>Application {descriptor.label}</h2>
      <p style={{ color: '#9ca3af' }}>This application is no longer active. Please contact support if you believe this is an error.</p>
    </div>
  );
}
