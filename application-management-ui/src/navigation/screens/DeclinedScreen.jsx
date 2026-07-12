export default function DeclinedScreen() {
  return (
    <div className="status-center" style={{ maxWidth: 520 }}>
      <div style={{ fontSize: '3rem', color: '#b71c1c', marginBottom: 16 }}>✗</div>
      <h2 style={{ color: '#b71c1c' }}>Application Declined</h2>
      <p style={{ color: '#6b7280' }}>
        Unfortunately, we are unable to approve your loan application at this time.
        You will receive a written notice with the reasons for this decision in accordance with the Fair Credit Reporting Act.
      </p>
      <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
        If you have questions, please contact our support team.
      </p>
    </div>
  );
}
