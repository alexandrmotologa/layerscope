import React, { useState, useEffect } from 'react';
import {
  Layers,
  ShieldAlert,
  Cpu,
  HardDrive,
  CheckCircle,
  RefreshCw,
  X,
} from 'lucide-react';
import { ImageData, SecurityReport, ImagePreset } from '../types';
import { api } from '../api/client';

interface HeaderProps {
  imageData: ImageData | null;
  securityData: SecurityReport | null;
  onImageChanged?: () => void;
}

export const Header: React.FC<HeaderProps> = ({ imageData, securityData, onImageChanged }) => {
  const [showModal, setShowModal] = useState(false);
  const [target, setTarget] = useState('');
  const [architecture, setArchitecture] = useState('amd64');
  const [forceRemote, setForceRemote] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [presets, setPresets] = useState<ImagePreset[]>([]);

  useEffect(() => {
    api
      .getPresets()
      .then((data) => setPresets(data))
      .catch((err) => console.error('Failed to load presets', err));
  }, []);

  const sizeMB = imageData ? (imageData.totalSizeBytes / (1024 * 1024)).toFixed(1) : '0.0';
  const arch = imageData ? `${imageData.reference.os}/${imageData.reference.architecture}` : 'linux/amd64';
  const effScore = imageData ? imageData.efficiencyScore.toFixed(1) : '100.0';
  const grade = imageData ? imageData.grade : 'A+';
  const interLeaks = securityData?.intermediateLeaks || 0;

  const handleAnalyze = async (imageTarget?: string, isDemoFlag = false) => {
    const finalTarget = imageTarget !== undefined ? imageTarget : target.trim();
    if (!finalTarget && !isDemoFlag) return;

    setAnalyzing(true);
    setError(null);
    try {
      await api.analyzeImage(finalTarget, forceRemote, architecture, isDemoFlag || finalTarget === 'demo');
      setShowModal(false);
      if (onImageChanged) {
        onImageChanged();
      }
    } catch (err: any) {
      setError(err.message || 'Failed to analyze target image');
    } finally {
      setAnalyzing(false);
    }
  };

  return (
    <>
      <header
        style={{
          background: 'rgba(13, 19, 31, 0.85)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '0.75rem 1.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.875rem' }}>
          <div
            style={{
              width: '34px',
              height: '34px',
              background: 'linear-gradient(135deg, #38bdf8, #6366f1)',
              borderRadius: '8px',
              display: 'grid',
              placeItems: 'center',
              color: 'white',
              boxShadow: '0 4px 12px rgba(56, 189, 248, 0.25)',
            }}
          >
            <Layers size={20} />
          </div>
          <div>
            <div
              style={{
                fontWeight: 800,
                fontSize: '1.25rem',
                letterSpacing: '-0.025em',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}
            >
              LayerScope
              <span
                style={{
                  fontSize: '0.7rem',
                  color: 'var(--accent)',
                  background: 'var(--accent-glow)',
                  padding: '0.15rem 0.5rem',
                  borderRadius: '9999px',
                  border: '1px solid rgba(56,189,248,0.3)',
                  fontWeight: 600,
                }}
              >
                v1.0.0
              </span>
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              OCI Container Layer Inspector & Security Auditor
            </div>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.625rem', fontSize: '0.8125rem' }}>
          <div
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border)',
              padding: '0.35rem 0.75rem',
              borderRadius: '9999px',
              display: 'flex',
              alignItems: 'center',
              gap: '0.45rem',
              fontFamily: 'JetBrains Mono, monospace',
            }}
          >
            <span style={{ color: 'var(--text-muted)' }}>Image:</span>
            <strong>{imageData?.reference.original || 'No image loaded'}</strong>
          </div>

          <div
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border)',
              padding: '0.35rem 0.75rem',
              borderRadius: '9999px',
              display: 'flex',
              alignItems: 'center',
              gap: '0.45rem',
              fontFamily: 'JetBrains Mono, monospace',
            }}
          >
            <Cpu size={14} color="var(--accent)" />
            <span>{arch}</span>
          </div>

          <div
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border)',
              padding: '0.35rem 0.75rem',
              borderRadius: '9999px',
              display: 'flex',
              alignItems: 'center',
              gap: '0.45rem',
              fontFamily: 'JetBrains Mono, monospace',
            }}
          >
            <HardDrive size={14} color="var(--text-muted)" />
            <span>{sizeMB} MB</span>
          </div>

          <div
            style={{
              background: 'rgba(5, 150, 105, 0.15)',
              border: '1px solid rgba(5, 150, 105, 0.4)',
              color: '#34d399',
              padding: '0.35rem 0.75rem',
              borderRadius: '9999px',
              display: 'flex',
              alignItems: 'center',
              gap: '0.45rem',
              fontFamily: 'JetBrains Mono, monospace',
              fontWeight: 700,
            }}
          >
            <CheckCircle size={14} />
            <span>
              {effScore}% ({grade})
            </span>
          </div>

          {interLeaks > 0 && (
            <div
              style={{
                background: 'rgba(220, 38, 38, 0.15)',
                border: '1px solid rgba(220, 38, 38, 0.4)',
                color: '#f87171',
                padding: '0.35rem 0.75rem',
                borderRadius: '9999px',
                display: 'flex',
                alignItems: 'center',
                gap: '0.45rem',
                fontFamily: 'JetBrains Mono, monospace',
                fontWeight: 700,
              }}
            >
              <ShieldAlert size={14} />
              <span>{interLeaks} Intermediate Leaks</span>
            </div>
          )}

          {/* Switch / Analyze Image Button */}
          <button
            onClick={() => setShowModal(true)}
            style={{
              background: 'linear-gradient(135deg, rgba(56, 189, 248, 0.2), rgba(99, 102, 241, 0.2))',
              border: '1px solid rgba(56, 189, 248, 0.4)',
              color: 'var(--accent)',
              padding: '0.4rem 0.85rem',
              borderRadius: '8px',
              fontWeight: 700,
              fontSize: '0.8125rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              transition: 'all 0.15s ease',
            }}
          >
            <RefreshCw size={13} /> Switch Image
          </button>
        </div>
      </header>

      {/* Switch / Load Image Modal */}
      {showModal && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(6px)',
            display: 'grid',
            placeItems: 'center',
            zIndex: 100,
            padding: '1rem',
          }}
        >
          <div
            style={{
              width: '100%',
              maxWidth: '560px',
              background: 'var(--bg-surface)',
              border: '1px solid var(--border)',
              borderRadius: '12px',
              boxShadow: '0 20px 40px rgba(0, 0, 0, 0.6)',
              overflow: 'hidden',
            }}
          >
            {/* Modal Header */}
            <div
              style={{
                padding: '1.25rem 1.5rem',
                borderBottom: '1px solid var(--border)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                background: 'var(--bg-card)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Layers size={20} color="var(--accent)" />
                <h3 style={{ fontSize: '1.1rem', fontWeight: 700, margin: 0, color: 'white' }}>
                  Analyze Container Image
                </h3>
              </div>
              <button
                onClick={() => setShowModal(false)}
                style={{
                  background: 'transparent',
                  border: 'none',
                  color: 'var(--text-muted)',
                  cursor: 'pointer',
                  padding: '0.25rem',
                }}
              >
                <X size={20} />
              </button>
            </div>

            {/* Modal Body */}
            <div style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
              {/* Presets Grid */}
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.5rem' }}>
                  Quick-Load Preset Images
                </label>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: '0.5rem' }}>
                  {presets.map((p) => (
                    <button
                      key={p.id}
                      onClick={() => handleAnalyze(p.target, p.isDemo === 'true')}
                      disabled={analyzing}
                      style={{
                        textAlign: 'left',
                        padding: '0.65rem 0.85rem',
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border)',
                        borderRadius: '6px',
                        cursor: 'pointer',
                        transition: 'border-color 0.15s ease',
                      }}
                    >
                      <div style={{ fontSize: '0.8125rem', fontWeight: 700, color: 'white', marginBottom: '0.2rem' }}>
                        {p.name}
                      </div>
                      <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', lineHeight: 1.3 }}>
                        {p.description}
                      </div>
                    </button>
                  ))}
                </div>
              </div>

              {/* Custom Input */}
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.5rem' }}>
                  Or Enter Custom Image Reference / Registry URI
                </label>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <input
                    type="text"
                    placeholder="e.g. alpine:3.20, node:20-alpine, or docker.io/..."
                    value={target}
                    onChange={(e) => setTarget(e.target.value)}
                    style={{
                      flex: 1,
                      background: 'var(--bg-card)',
                      border: '1px solid var(--border)',
                      padding: '0.55rem 0.875rem',
                      borderRadius: '6px',
                      color: 'white',
                      fontSize: '0.875rem',
                      outline: 'none',
                    }}
                  />
                  <select
                    value={architecture}
                    onChange={(e) => setArchitecture(e.target.value)}
                    style={{
                      background: 'var(--bg-card)',
                      border: '1px solid var(--border)',
                      color: 'var(--text)',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      fontSize: '0.8125rem',
                      outline: 'none',
                      cursor: 'pointer',
                    }}
                  >
                    <option value="amd64">amd64</option>
                    <option value="arm64">arm64</option>
                  </select>
                </div>
              </div>

              {/* Options */}
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8125rem', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={forceRemote}
                  onChange={(e) => setForceRemote(e.target.checked)}
                />
                <span>Fetch directly from remote registry (bypass local Docker daemon)</span>
              </label>

              {/* Error Message */}
              {error && (
                <div style={{ padding: '0.75rem', borderRadius: '6px', background: 'rgba(239, 68, 68, 0.15)', color: '#ef4444', fontSize: '0.8125rem' }}>
                  {error}
                </div>
              )}

              {/* Actions */}
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
                <button
                  onClick={() => setShowModal(false)}
                  disabled={analyzing}
                  style={{
                    padding: '0.5rem 1rem',
                    borderRadius: '6px',
                    border: '1px solid var(--border)',
                    background: 'transparent',
                    color: 'var(--text-muted)',
                    fontSize: '0.8125rem',
                    cursor: 'pointer',
                  }}
                >
                  Cancel
                </button>
                <button
                  onClick={() => handleAnalyze()}
                  disabled={analyzing || !target.trim()}
                  style={{
                    padding: '0.5rem 1.25rem',
                    borderRadius: '6px',
                    border: 'none',
                    background: 'var(--accent)',
                    color: '#0f172a',
                    fontWeight: 700,
                    fontSize: '0.8125rem',
                    cursor: analyzing || !target.trim() ? 'not-allowed' : 'pointer',
                    opacity: analyzing || !target.trim() ? 0.6 : 1,
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.4rem',
                  }}
                >
                  {analyzing ? 'Analyzing Layers...' : 'Analyze Image'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
