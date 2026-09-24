import React, { useState, useMemo } from 'react';
import { Package, Download, Search } from 'lucide-react';
import { SBOMReport } from '../types';

interface SbomViewerProps {
  report: SBOMReport | null;
}

export const SbomViewer: React.FC<SbomViewerProps> = ({ report }) => {
  const [search, setSearch] = useState('');

  const filteredPackages = useMemo(() => {
    if (!report?.packages) return [];
    const q = search.toLowerCase().trim();
    return report.packages.filter((p) => {
      if (!q) return true;
      return (
        p.name.toLowerCase().includes(q) ||
        p.version.toLowerCase().includes(q) ||
        (p.license && p.license.toLowerCase().includes(q)) ||
        (p.purl && p.purl.toLowerCase().includes(q))
      );
    });
  }, [report, search]);

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem', background: 'var(--bg-dark)' }}>
      {/* Header bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
            <Package size={18} color="var(--accent)" />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Software Bill of Materials (SBOM)</h3>
          </div>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
            {report ? `${report.totalPackages} components identified (${report.osPackages} OS packages, ${report.appPackages} runtime dependencies)` : 'Analyzing image...'}
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            onClick={() => window.open('/api/sbom/cyclonedx')}
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              color: 'var(--text)',
              padding: '0.45rem 0.875rem',
              borderRadius: '6px',
              fontSize: '0.8125rem',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
            }}
          >
            <Download size={14} /> CycloneDX 1.5 JSON
          </button>
          <button
            onClick={() => window.open('/api/sbom/spdx')}
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              color: 'var(--text)',
              padding: '0.45rem 0.875rem',
              borderRadius: '6px',
              fontSize: '0.8125rem',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
            }}
          >
            <Download size={14} /> SPDX 2.3 JSON
          </button>
        </div>
      </div>

      {/* Search Input */}
      <div style={{
        position: 'relative',
        marginBottom: '1rem',
        display: 'flex',
        alignItems: 'center',
      }}>
        <Search size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.875rem' }} />
        <input
          type="text"
          placeholder="Filter package name, version, license..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{
            width: '100%',
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            padding: '0.5rem 1rem 0.5rem 2.5rem',
            borderRadius: '6px',
            color: 'white',
            fontSize: '0.875rem',
            outline: 'none',
          }}
        />
      </div>

      {/* Packages Table */}
      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8125rem' }}>
          <thead>
            <tr style={{ background: 'var(--bg-surface)', borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
              <th style={{ padding: '0.75rem 1rem' }}>Ecosystem</th>
              <th style={{ padding: '0.75rem 1rem' }}>Package Name</th>
              <th style={{ padding: '0.75rem 1rem' }}>Version</th>
              <th style={{ padding: '0.75rem 1rem' }}>License</th>
              <th style={{ padding: '0.75rem 1rem' }}>Package URL (PURL)</th>
            </tr>
          </thead>
          <tbody>
            {filteredPackages.map((pkg, idx) => (
              <tr key={idx} style={{ borderBottom: '1px solid rgba(30,41,59,0.5)' }}>
                <td style={{ padding: '0.65rem 1rem' }}>
                  <span style={{
                    padding: '0.15rem 0.45rem',
                    borderRadius: '4px',
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    background: pkg.type.startsWith('os') ? 'rgba(56,189,248,0.15)' : 'rgba(168,85,247,0.15)',
                    color: pkg.type.startsWith('os') ? 'var(--accent)' : 'var(--purple)',
                  }}>
                    {pkg.type}
                  </span>
                </td>
                <td style={{ padding: '0.65rem 1rem', fontWeight: 600 }}>{pkg.name}</td>
                <td style={{ padding: '0.65rem 1rem', fontFamily: 'JetBrains Mono, monospace' }}>{pkg.version}</td>
                <td style={{ padding: '0.65rem 1rem', color: 'var(--text-muted)' }}>{pkg.license || 'N/A'}</td>
                <td style={{ padding: '0.65rem 1rem', fontFamily: 'JetBrains Mono, monospace', color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                  <code>{pkg.purl}</code>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
