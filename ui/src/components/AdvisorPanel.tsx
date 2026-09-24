import React from 'react';
import { Sparkles, TrendingDown, Terminal } from 'lucide-react';
import { AdvisorReport } from '../types';

interface AdvisorPanelProps {
  report: AdvisorReport | null;
}

export const AdvisorPanel: React.FC<AdvisorPanelProps> = ({ report }) => {
  if (!report) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
        Generating optimization recommendations...
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem', background: 'var(--bg-dark)' }}>
      {/* Banner */}
      <div style={{
        background: 'linear-gradient(135deg, rgba(30, 58, 138, 0.4), rgba(15, 23, 42, 0.8))',
        border: '1px solid rgba(59, 130, 246, 0.3)',
        borderRadius: '8px',
        padding: '1.25rem 1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '1.5rem',
      }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
            <Sparkles size={18} color="var(--accent)" />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 800 }}>Image Efficiency Score: {report.efficiencyScore.toFixed(1)}%</h3>
            <span style={{
              background: '#064e3b',
              color: '#34d399',
              padding: '0.15rem 0.6rem',
              borderRadius: '9999px',
              fontWeight: 800,
              fontSize: '0.8125rem',
            }}>
              Grade {report.grade}
            </span>
          </div>
          <p style={{ color: '#cbd5e1', fontSize: '0.8125rem' }}>
            {report.recommendations.length} optimization opportunities detected to eliminate build cache busts and layer bloat.
          </p>
        </div>

        <div style={{
          textAlign: 'right',
          background: 'rgba(15, 23, 42, 0.6)',
          padding: '0.5rem 1rem',
          borderRadius: '6px',
          border: '1px solid var(--border)',
        }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Wasted Payload Space</div>
          <div style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--yellow)', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <TrendingDown size={16} />
            {(report.totalWastedBytes / (1024 * 1024)).toFixed(2)} MB
          </div>
        </div>
      </div>

      {/* Recommendations Cards */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {report.recommendations.map((rec) => (
          <div
            key={rec.id}
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              padding: '1.25rem',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
              <div>
                <h4 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--text)', marginBottom: '0.25rem' }}>
                  {rec.title}
                </h4>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'flex', gap: '0.75rem' }}>
                  <span>Category: <strong>{rec.category}</strong></span>
                  {rec.estimatedSavingsBytes > 0 && (
                    <span style={{ color: 'var(--green)' }}>
                      Est. Savings: <strong>{(rec.estimatedSavingsBytes / 1024).toFixed(0)} KB</strong>
                    </span>
                  )}
                </div>
              </div>

              <span style={{
                padding: '0.15rem 0.5rem',
                borderRadius: '4px',
                fontSize: '0.7rem',
                fontWeight: 800,
                textTransform: 'uppercase',
                background: rec.severity === 'high' ? 'rgba(239,68,68,0.15)' : 'rgba(234,179,8,0.15)',
                color: rec.severity === 'high' ? 'var(--red)' : 'var(--yellow)',
              }}>
                {rec.severity}
              </span>
            </div>

            <p style={{ color: '#cbd5e1', fontSize: '0.875rem', lineHeight: 1.5, margin: '0.75rem 0' }}>
              {rec.explanation}
            </p>

            {rec.remediationSnippet && (
              <div>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.35rem',
                  fontSize: '0.75rem',
                  color: 'var(--text-muted)',
                  marginBottom: '0.25rem',
                }}>
                  <Terminal size={13} />
                  <span>Actionable Dockerfile Remediation:</span>
                </div>
                <pre style={{
                  background: '#050811',
                  border: '1px solid #1e293b',
                  borderRadius: '6px',
                  padding: '0.875rem 1rem',
                  fontSize: '0.8125rem',
                  color: 'var(--accent)',
                  lineHeight: 1.4,
                  overflowX: 'auto',
                }}>
                  <code>{rec.remediationSnippet}</code>
                </pre>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
