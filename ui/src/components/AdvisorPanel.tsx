import React, { useState, useEffect } from 'react';
import { Sparkles, TrendingDown, Terminal, Copy, Check, FileCode } from 'lucide-react';
import { AdvisorReport, DockerfileOptimization } from '../types';
import { api } from '../api/client';

interface AdvisorPanelProps {
  report: AdvisorReport | null;
}

export const AdvisorPanel: React.FC<AdvisorPanelProps> = ({ report }) => {
  const [optimization, setOptimization] = useState<DockerfileOptimization | null>(null);
  const [activeTab, setActiveTab] = useState<'dockerfile' | 'dockerignore'>('dockerfile');
  const [copiedFile, setCopiedFile] = useState<string | null>(null);

  useEffect(() => {
    api
      .getDockerfileOptimization()
      .then((data) => setOptimization(data))
      .catch((err) => console.error('Failed to load dockerfile optimization', err));
  }, [report]);

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopiedFile(label);
    setTimeout(() => setCopiedFile(null), 2000);
  };

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
      <div
        style={{
          background: 'linear-gradient(135deg, rgba(30, 58, 138, 0.4), rgba(15, 23, 42, 0.8))',
          border: '1px solid rgba(59, 130, 246, 0.3)',
          borderRadius: '8px',
          padding: '1.25rem 1.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1.5rem',
        }}
      >
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
            <Sparkles size={18} color="var(--accent)" />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 800 }}>
              Image Efficiency Score: {report.efficiencyScore.toFixed(1)}%
            </h3>
            <span
              style={{
                background: '#064e3b',
                color: '#34d399',
                padding: '0.15rem 0.6rem',
                borderRadius: '9999px',
                fontWeight: 800,
                fontSize: '0.8125rem',
              }}
            >
              Grade {report.grade}
            </span>
          </div>
          <p style={{ color: '#cbd5e1', fontSize: '0.8125rem', margin: 0 }}>
            {report.recommendations.length} optimization opportunities detected to eliminate build cache busts and layer bloat.
          </p>
        </div>

        <div
          style={{
            textAlign: 'right',
            background: 'rgba(15, 23, 42, 0.6)',
            padding: '0.5rem 1rem',
            borderRadius: '6px',
            border: '1px solid var(--border)',
          }}
        >
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Wasted Payload Space</div>
          <div
            style={{
              fontSize: '1.25rem',
              fontWeight: 800,
              color: 'var(--yellow)',
              display: 'flex',
              alignItems: 'center',
              gap: '0.25rem',
            }}
          >
            <TrendingDown size={16} />
            {(report.totalWastedBytes / (1024 * 1024)).toFixed(2)} MB
          </div>
        </div>
      </div>

      {/* Dockerfile Auto-Fixer & Optimizer View */}
      {optimization && (
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid rgba(56, 189, 248, 0.3)',
            borderRadius: '8px',
            marginBottom: '1.5rem',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '1rem 1.25rem',
              background: 'rgba(56, 189, 248, 0.05)',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <FileCode size={18} color="var(--accent)" />
                <h4 style={{ fontSize: '1rem', fontWeight: 700, margin: 0, color: 'white' }}>
                  Auto-Optimized Dockerfile & .dockerignore Generator
                </h4>
                <span
                  style={{
                    background: 'rgba(34, 197, 94, 0.15)',
                    color: 'var(--emerald)',
                    padding: '0.15rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.725rem',
                    fontWeight: 700,
                  }}
                >
                  ~{optimization.estimatedSavingsMB.toFixed(1)} MB Estimated Savings
                </span>
              </div>
              <p style={{ color: 'var(--text-muted)', fontSize: '0.775rem', margin: '0.25rem 0 0 0' }}>
                Production-hardened, cache-optimized Docker build configuration tailored to this container's runtime.
              </p>
            </div>

            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <div
                style={{
                  display: 'flex',
                  background: 'var(--bg-dark)',
                  padding: '0.2rem',
                  borderRadius: '6px',
                  border: '1px solid var(--border)',
                }}
              >
                <button
                  onClick={() => setActiveTab('dockerfile')}
                  style={{
                    padding: '0.3rem 0.75rem',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    borderRadius: '4px',
                    border: 'none',
                    background: activeTab === 'dockerfile' ? 'var(--accent)' : 'transparent',
                    color: activeTab === 'dockerfile' ? '#0f172a' : 'var(--text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  Dockerfile.optimized
                </button>
                <button
                  onClick={() => setActiveTab('dockerignore')}
                  style={{
                    padding: '0.3rem 0.75rem',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    borderRadius: '4px',
                    border: 'none',
                    background: activeTab === 'dockerignore' ? 'var(--accent)' : 'transparent',
                    color: activeTab === 'dockerignore' ? '#0f172a' : 'var(--text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  .dockerignore
                </button>
              </div>

              <button
                onClick={() =>
                  handleCopy(
                    activeTab === 'dockerfile' ? optimization.optimizedDockerfile : optimization.generatedDockerignore,
                    activeTab
                  )
                }
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.35rem',
                  background: 'var(--accent)',
                  color: '#0f172a',
                  border: 'none',
                  padding: '0.4rem 0.8rem',
                  borderRadius: '6px',
                  fontSize: '0.75rem',
                  fontWeight: 700,
                  cursor: 'pointer',
                }}
              >
                {copiedFile === activeTab ? <Check size={13} /> : <Copy size={13} />}
                <span>{copiedFile === activeTab ? 'Copied!' : `Copy ${activeTab === 'dockerfile' ? 'Dockerfile' : '.dockerignore'}`}</span>
              </button>
            </div>
          </div>

          {/* Key Improvements Highlights */}
          {optimization.improvements.length > 0 && (
            <div
              style={{
                padding: '0.75rem 1.25rem',
                borderBottom: '1px solid var(--border)',
                background: 'rgba(15, 23, 42, 0.4)',
                fontSize: '0.75rem',
              }}
            >
              <div style={{ fontWeight: 700, color: 'var(--accent)', marginBottom: '0.35rem' }}>
                Key Architectural Remediations Applied:
              </div>
              <ul style={{ margin: 0, paddingLeft: '1.25rem', color: '#cbd5e1', lineHeight: 1.5 }}>
                {optimization.improvements.map((imp, idx) => (
                  <li key={idx}>{imp}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Code Viewer */}
          <pre
            style={{
              margin: 0,
              background: '#090d16',
              padding: '1.25rem',
              fontFamily: 'JetBrains Mono, monospace',
              fontSize: '0.8rem',
              lineHeight: 1.5,
              color: '#38bdf8',
              overflowX: 'auto',
            }}
          >
            <code>
              {activeTab === 'dockerfile' ? optimization.optimizedDockerfile : optimization.generatedDockerignore}
            </code>
          </pre>
        </div>
      )}

      {/* Recommendations Cards Header */}
      <h4 style={{ fontSize: '0.95rem', fontWeight: 700, marginBottom: '0.75rem', color: 'var(--text-muted)' }}>
        Detailed Heuristic Analysis & Optimization Rules
      </h4>

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
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: '0.5rem',
              }}
            >
              <div>
                <h4 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--text)', marginBottom: '0.25rem' }}>
                  {rec.title}
                </h4>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'flex', gap: '0.75rem' }}>
                  <span>
                    Category: <strong>{rec.category}</strong>
                  </span>
                  {rec.estimatedSavingsBytes > 0 && (
                    <span style={{ color: 'var(--green)' }}>
                      Est. Savings: <strong>{(rec.estimatedSavingsBytes / 1024).toFixed(0)} KB</strong>
                    </span>
                  )}
                </div>
              </div>

              <span
                style={{
                  padding: '0.15rem 0.5rem',
                  borderRadius: '4px',
                  fontSize: '0.7rem',
                  fontWeight: 800,
                  textTransform: 'uppercase',
                  background: rec.severity === 'high' ? 'rgba(239,68,68,0.15)' : 'rgba(234,179,8,0.15)',
                  color: rec.severity === 'high' ? 'var(--red)' : 'var(--yellow)',
                }}
              >
                {rec.severity}
              </span>
            </div>

            <p style={{ color: '#cbd5e1', fontSize: '0.875rem', lineHeight: 1.5, margin: '0.75rem 0' }}>
              {rec.explanation}
            </p>

            {rec.remediationSnippet && (
              <div>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.35rem',
                    fontSize: '0.75rem',
                    color: 'var(--text-muted)',
                    marginBottom: '0.25rem',
                  }}
                >
                  <Terminal size={13} />
                  <span>Actionable Dockerfile Remediation:</span>
                </div>
                <pre
                  style={{
                    background: '#050811',
                    border: '1px solid #1e293b',
                    borderRadius: '6px',
                    padding: '0.875rem 1rem',
                    fontSize: '0.8125rem',
                    color: 'var(--accent)',
                    lineHeight: 1.4,
                    overflowX: 'auto',
                  }}
                >
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
