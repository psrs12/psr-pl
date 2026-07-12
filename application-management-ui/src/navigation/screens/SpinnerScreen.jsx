import React from 'react';

export default function SpinnerScreen({ descriptor }) {
  return (
    <div className="status-center">
      <div className="spinner" />
      <h3 style={{ color: '#1a1a1a' }}>{descriptor.label}</h3>
      <p style={{ color: '#6b7280' }}>
        {descriptor.description.split('\n').map((line, i) => (
          <React.Fragment key={i}>
            {i > 0 && <br />}
            {line}
          </React.Fragment>
        ))}
      </p>
    </div>
  );
}
