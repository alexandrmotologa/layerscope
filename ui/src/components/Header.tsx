import React from 'react';
import { Layers, ShieldAlert, Cpu, HardDrive, CheckCircle } from 'lucide-react';
import { ImageData, SecurityReport } from '../types';

interface HeaderProps {
  imageData: ImageData | null;
  securityData: SecurityReport | null;
}

export const Header: React.FC<HeaderProps> = ({ imageData, securityData }) => {
  const sizeMB = imageData ? (imageData.totalSizeBytes / (1024 * 1024)).toFixed(1) : '0.0';
  const arch = imageData ? `${imageData.reference.os}/${imageData.reference.architecture}` : 'linux/amd64';
  const effScore = imageData ? imageData.efficiencyScore.toFixed(1) : '100.0';
  const grade = imageData ? imageData.grade : 'A+';
  const interLeaks = securityData?.intermediateLeaks || 0;

  return (
    <header style={{
      background: 'rgba(13, 19, 31, 0.85)',
      backdropFilter: 'blur(12px)',
      borderBottom: '1px solid var(--border)',
      padding: '0.75rem 1.5rem',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.875rem' }}>
        <div style={{
          width: '34px',
          height: '34px',
          background: 'linear-gradient(135deg, #38bdf8, #6366f1)',
          borderRadius: '8px',
          display: 'grid',
          placeItems: 'center',
          color: 'white',
          boxShadow: '0 4px 12px rgba(56, 189, 248, 0.25)',
        }}>
          <Layers size={20} />
        </div>
        <div>
          <div style={{ fontWeight: 800, fontSize: '1.25rem', letterSpacing: '-0.025em', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            LayerScope
            <span style={{ fontSize: '0.7rem', color: 'var(--accent)', background: 'var(--accent-glow)', padding: '0.15rem 0.5rem', borderRadius: '9999px', border: '1px solid rgba(56,189,248,0.3)', fontWeight: 600 }}>v1.0.0</span>
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>OCI Container Layer Inspector & Security Auditor</div>
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '0.625rem', fontSize: '0.8125rem' }}>
        <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', padding: '0.35rem 0.75rem', borderRadius: '9999px', display: 'flex', alignItems: 'center', gap: '0.45rem', fontFamily: 'JetBrains Mono, monospace' }}>
          <span style={{ color: 'var(--text-muted)' }}>Image:</span>
          <strong>{imageData?.reference.original || 'No image loaded'}</strong>
        </div>

        <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', padding: '0.35rem 0.75rem', borderRadius: '9999px', display: 'flex', alignItems: 'center', gap: '0.45rem', fontFamily: 'JetBrains Mono, monospace' }}>
          <Cpu size={14} color="var(--accent)" />
          <span>{arch}</span>
        </div>

        <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', padding: '0.35rem 0.75rem', borderRadius: '9999px', display: 'flex', alignItems: 'center', gap: '0.45rem', fontFamily: 'JetBrains Mono, monospace' }}>
          <HardDrive size={14} color="var(--text-muted)" />
          <span>{sizeMB} MB</span>
        </div>

        <div style={{
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
        }}>
          <CheckCircle size={14} />
          <span>{effScore}% ({grade})</span>
        </div>

        {interLeaks > 0 && (
          <div style={{
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
          }}>
            <ShieldAlert size={14} />
            <span>{interLeaks} Intermediate Leaks</span>
          </div>
        )}
      </div>
    </header>
  );
};
