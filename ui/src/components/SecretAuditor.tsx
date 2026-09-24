import React from 'react';
import { ShieldAlert, AlertOctagon, Key, UserCheck, Lock } from 'lucide-react';
import { SecurityReport } from '../types';

interface SecretAuditorProps {
  report: SecurityReport | null;
}

export const SecretAuditor: React.FC<SecretAuditorProps> = ({ report }) => {
  if (!report) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
        Loading security audit...
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem', background: 'var(--bg-dark)' }}>
      {/* Metrics Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '1rem', marginBottom: '1.5rem' }}>
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1.25rem', borderRadius: '8px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>Total Credentials Leaked</span>
            <Key size={16} color="var(--red)" />
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: report.totalFindings > 0 ? 'var(--red)' : 'var(--green)' }}>
            {report.totalFindings}
          </div>
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1.25rem', borderRadius: '8px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>Intermediate Layer Leaks</span>
            <AlertOctagon size={16} color="#f87171" />
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: report.intermediateLeaks > 0 ? '#f87171' : 'var(--green)' }}>
            {report.intermediateLeaks}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Secrets deleted in subsequent layers
          </div>
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1.25rem', borderRadius: '8px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>Default Container User</span>
            <UserCheck size={16} color={report.runsAsRoot ? 'var(--red)' : 'var(--green)'} />
          </div>
          <div style={{ fontSize: '1.25rem', fontWeight: 800, color: report.runsAsRoot ? 'var(--red)' : 'var(--green)', marginTop: '0.35rem' }}>
            {report.runsAsRoot ? 'Root (UID 0 - Risk)' : 'Non-Root User'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1.25rem', borderRadius: '8px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>SUID / SGID Binaries</span>
            <Lock size={16} color="var(--yellow)" />
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--text)' }}>
            {report.suidBinariesCount}
          </div>
        </div>
      </div>

      {/* Findings Table */}
      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
        <div style={{ padding: '0.875rem 1rem', borderBottom: '1px solid var(--border)', fontWeight: 700, fontSize: '0.9rem' }}>
          Detected Credentials Across Historical Layers
        </div>

        {report.findings.length === 0 ? (
          <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--green)' }}>
            No credentials or sensitive tokens detected in any image layer.
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8125rem' }}>
            <thead>
              <tr style={{ background: 'var(--bg-surface)', borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                <th style={{ padding: '0.75rem 1rem' }}>Severity</th>
                <th style={{ padding: '0.75rem 1rem' }}>Rule Name</th>
                <th style={{ padding: '0.75rem 1rem' }}>File Location</th>
                <th style={{ padding: '0.75rem 1rem' }}>Layer</th>
                <th style={{ padding: '0.75rem 1rem' }}>Intermediate Leak</th>
                <th style={{ padding: '0.75rem 1rem' }}>Masked Credential</th>
              </tr>
            </thead>
            <tbody>
              {report.findings.map((f) => (
                <tr key={f.id} style={{ borderBottom: '1px solid rgba(30,41,59,0.5)' }}>
                  <td style={{ padding: '0.75rem 1rem' }}>
                    <span style={{
                      padding: '0.15rem 0.45rem',
                      borderRadius: '4px',
                      fontWeight: 800,
                      fontSize: '0.7rem',
                      background: f.severity === 'CRITICAL' ? 'rgba(239,68,68,0.2)' : 'rgba(234,179,8,0.2)',
                      color: f.severity === 'CRITICAL' ? 'var(--red)' : 'var(--yellow)',
                    }}>
                      {f.severity}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>{f.ruleName}</td>
                  <td style={{ padding: '0.75rem 1rem', fontFamily: 'JetBrains Mono, monospace' }}>
                    {f.filePath}:{f.lineNumber}
                  </td>
                  <td style={{ padding: '0.75rem 1rem' }}>Layer {f.layerIndex}</td>
                  <td style={{ padding: '0.75rem 1rem' }}>
                    {f.isDeletedInFinal ? (
                      <span style={{ color: '#f87171', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                        <ShieldAlert size={14} /> YES (Deleted in later layer)
                      </span>
                    ) : (
                      <span style={{ color: 'var(--text-muted)' }}>No (Present in final)</span>
                    )}
                  </td>
                  <td style={{ padding: '0.75rem 1rem', fontFamily: 'JetBrains Mono, monospace', color: 'var(--accent)' }}>
                    <code>{f.matchMasked}</code>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
