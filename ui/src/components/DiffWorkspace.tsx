import React, { useState } from 'react';
import { GitCompare, RefreshCw } from 'lucide-react';
import { api } from '../api/client';
import { DiffReport } from '../types';

interface DiffWorkspaceProps {
  currentImageName: string;
}

export const DiffWorkspace: React.FC<DiffWorkspaceProps> = ({ currentImageName }) => {
  const [imgA, setImgA] = useState(currentImageName);
  const [imgB, setImgB] = useState('');
  const [archA, setArchA] = useState('amd64');
  const [archB, setArchB] = useState('arm64');
  const [loading, setLoading] = useState(false);
  const [diffReport, setDiffReport] = useState<DiffReport | null>(null);
  const [error, setError] = useState('');

  const handleDiff = async () => {
    if (!imgB) {
      setError('Please provide a second image or architecture target to compare.');
      return;
    }
    setError('');
    setLoading(true);
    try {
      const res = await api.diff(imgA, imgB, archA, archB);
      setDiffReport(res);
    } catch (err: any) {
      setError(err.message || 'Failed to calculate diff');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem', background: 'var(--bg-dark)' }}>
      {/* Input controls */}
      <div style={{
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: '8px',
        padding: '1.25rem',
        marginBottom: '1.5rem',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
          <GitCompare size={18} color="var(--accent)" />
          <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>Side-by-Side Tag & Multi-Arch Comparator</h3>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: '1rem', alignItems: 'flex-end' }}>
          <div>
            <label style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>
              Base Image (A)
            </label>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <input
                type="text"
                value={imgA}
                onChange={(e) => setImgA(e.target.value)}
                style={{
                  flex: 1,
                  background: 'var(--bg-surface)',
                  border: '1px solid var(--border)',
                  padding: '0.45rem 0.75rem',
                  borderRadius: '6px',
                  color: 'white',
                  fontFamily: 'JetBrains Mono, monospace',
                  fontSize: '0.8125rem',
                }}
              />
              <select
                value={archA}
                onChange={(e) => setArchA(e.target.value)}
                style={{
                  background: 'var(--bg-surface)',
                  border: '1px solid var(--border)',
                  padding: '0.45rem 0.5rem',
                  borderRadius: '6px',
                  color: 'white',
                  fontSize: '0.8125rem',
                }}
              >
                <option value="amd64">amd64</option>
                <option value="arm64">arm64</option>
              </select>
            </div>
          </div>

          <div>
            <label style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>
              Comparison Image (B)
            </label>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <input
                type="text"
                placeholder="e.g. myapp:v1.1.0 or demo"
                value={imgB}
                onChange={(e) => setImgB(e.target.value)}
                style={{
                  flex: 1,
                  background: 'var(--bg-surface)',
                  border: '1px solid var(--border)',
                  padding: '0.45rem 0.75rem',
                  borderRadius: '6px',
                  color: 'white',
                  fontFamily: 'JetBrains Mono, monospace',
                  fontSize: '0.8125rem',
                }}
              />
              <select
                value={archB}
                onChange={(e) => setArchB(e.target.value)}
                style={{
                  background: 'var(--bg-surface)',
                  border: '1px solid var(--border)',
                  padding: '0.45rem 0.5rem',
                  borderRadius: '6px',
                  color: 'white',
                  fontSize: '0.8125rem',
                }}
              >
                <option value="arm64">arm64</option>
                <option value="amd64">amd64</option>
              </select>
            </div>
          </div>

          <button
            onClick={handleDiff}
            disabled={loading}
            style={{
              background: 'linear-gradient(135deg, #2563eb, #3b82f6)',
              color: 'white',
              border: 'none',
              padding: '0.5rem 1.25rem',
              borderRadius: '6px',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
            }}
          >
            {loading ? <RefreshCw size={14} className="animate-spin" /> : <GitCompare size={14} />}
            Compare
          </button>
        </div>

        {error && (
          <div style={{ color: 'var(--red)', fontSize: '0.8125rem', marginTop: '0.75rem' }}>
            {error}
          </div>
        )}
      </div>

      {/* Results View */}
      {diffReport && (
        <div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '1rem', marginBottom: '1.5rem' }}>
            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1rem', borderRadius: '8px' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Net Delta (B - A)</div>
              <div style={{
                fontSize: '1.5rem',
                fontWeight: 800,
                color: diffReport.totalSizeDelta > 0 ? 'var(--red)' : 'var(--green)',
                marginTop: '0.25rem',
              }}>
                {diffReport.totalSizeDelta > 0 ? '+' : ''}{(diffReport.totalSizeDelta / (1024 * 1024)).toFixed(2)} MB
              </div>
            </div>

            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1rem', borderRadius: '8px' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Added Files</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--green)', marginTop: '0.25rem' }}>
                +{diffReport.addedCount}
              </div>
            </div>

            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1rem', borderRadius: '8px' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Removed Files</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--red)', marginTop: '0.25rem' }}>
                -{diffReport.removedCount}
              </div>
            </div>

            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', padding: '1rem', borderRadius: '8px' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Modified Files</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--yellow)', marginTop: '0.25rem' }}>
                ~{diffReport.modifiedCount}
              </div>
            </div>
          </div>

          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8125rem' }}>
              <thead>
                <tr style={{ background: 'var(--bg-surface)', borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                  <th style={{ padding: '0.75rem 1rem' }}>Status</th>
                  <th style={{ padding: '0.75rem 1rem' }}>File Path</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Delta</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Size in B</th>
                </tr>
              </thead>
              <tbody>
                {diffReport.files.slice(0, 100).map((f) => (
                  <tr key={f.path} style={{ borderBottom: '1px solid rgba(30,41,59,0.5)' }}>
                    <td style={{ padding: '0.6rem 1rem' }}>
                      <span style={{
                        padding: '0.15rem 0.45rem',
                        borderRadius: '4px',
                        fontWeight: 700,
                        fontSize: '0.7rem',
                        background: f.status === 'added' ? 'rgba(34,197,94,0.15)' : f.status === 'removed' ? 'rgba(239,68,68,0.15)' : 'rgba(234,179,8,0.15)',
                        color: f.status === 'added' ? 'var(--green)' : f.status === 'removed' ? 'var(--red)' : 'var(--yellow)',
                      }}>
                        {f.status}
                      </span>
                    </td>
                    <td style={{ padding: '0.6rem 1rem', fontFamily: 'JetBrains Mono, monospace' }}>{f.path}</td>
                    <td style={{ padding: '0.6rem 1rem', fontFamily: 'JetBrains Mono, monospace' }}>
                      {(f.sizeDelta / 1024).toFixed(1)} KB
                    </td>
                    <td style={{ padding: '0.6rem 1rem', fontFamily: 'JetBrains Mono, monospace', color: 'var(--text-muted)' }}>
                      {(f.sizeB / 1024).toFixed(1)} KB
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
