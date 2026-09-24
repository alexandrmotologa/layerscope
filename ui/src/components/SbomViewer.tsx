import React, { useState, useMemo } from 'react';
import { Package, Download, Search, ShieldAlert, ShieldCheck, ChevronDown, ChevronRight, ExternalLink } from 'lucide-react';
import { SBOMReport } from '../types';

interface SbomViewerProps {
  report: SBOMReport | null;
}

export const SbomViewer: React.FC<SbomViewerProps> = ({ report }) => {
  const [search, setSearch] = useState('');
  const [filterMode, setFilterMode] = useState<'all' | 'vulns' | 'os' | 'app'>('all');
  const [expandedPkg, setExpandedPkg] = useState<string | null>(null);

  const vulnCount = useMemo(() => {
    if (!report?.packages) return 0;
    return report.packages.filter((p) => p.vulnerabilities && p.vulnerabilities.length > 0).length;
  }, [report]);

  const filteredPackages = useMemo(() => {
    if (!report?.packages) return [];
    const q = search.toLowerCase().trim();
    return report.packages.filter((p) => {
      // Filter mode
      if (filterMode === 'vulns' && (!p.vulnerabilities || p.vulnerabilities.length === 0)) {
        return false;
      }
      if (filterMode === 'os' && !p.type.startsWith('os')) {
        return false;
      }
      if (filterMode === 'app' && p.type.startsWith('os')) {
        return false;
      }

      // Query
      if (!q) return true;
      const matchesText =
        p.name.toLowerCase().includes(q) ||
        p.version.toLowerCase().includes(q) ||
        (p.license && p.license.toLowerCase().includes(q)) ||
        (p.purl && p.purl.toLowerCase().includes(q));

      const matchesVuln =
        p.vulnerabilities?.some(
          (v) =>
            v.id.toLowerCase().includes(q) ||
            v.summary.toLowerCase().includes(q) ||
            v.aliases?.some((a) => a.toLowerCase().includes(q))
        ) || false;

      return matchesText || matchesVuln;
    });
  }, [report, search, filterMode]);

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem', background: 'var(--bg-dark)' }}>
      {/* Header bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
            <Package size={18} color="var(--accent)" />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Software Bill of Materials (SBOM) & CVE Tracker</h3>
          </div>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
            {report
              ? `${report.totalPackages} components identified (${report.osPackages} OS packages, ${report.appPackages} runtime packages) · Real-time OSV.dev matching`
              : 'Analyzing image...'}
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

      {/* Filter and Search Bar */}
      <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '1rem', alignItems: 'center' }}>
        <div style={{ position: 'relative', flex: 1, display: 'flex', alignItems: 'center' }}>
          <Search size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.875rem' }} />
          <input
            type="text"
            placeholder="Search package name, version, license, CVE ID..."
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

        <div style={{ display: 'flex', gap: '0.25rem', background: 'var(--bg-card)', padding: '0.25rem', borderRadius: '6px', border: '1px solid var(--border)' }}>
          <button
            onClick={() => setFilterMode('all')}
            style={{
              padding: '0.35rem 0.75rem',
              fontSize: '0.75rem',
              fontWeight: 600,
              borderRadius: '4px',
              border: 'none',
              background: filterMode === 'all' ? 'var(--accent)' : 'transparent',
              color: filterMode === 'all' ? '#0f172a' : 'var(--text-muted)',
              cursor: 'pointer',
            }}
          >
            All ({report?.totalPackages || 0})
          </button>
          <button
            onClick={() => setFilterMode('vulns')}
            style={{
              padding: '0.35rem 0.75rem',
              fontSize: '0.75rem',
              fontWeight: 600,
              borderRadius: '4px',
              border: 'none',
              background: filterMode === 'vulns' ? 'rgba(239, 68, 68, 0.2)' : 'transparent',
              color: filterMode === 'vulns' ? '#ef4444' : (vulnCount > 0 ? '#f87171' : 'var(--text-muted)'),
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.35rem',
            }}
          >
            <ShieldAlert size={12} /> Vulnerable ({vulnCount})
          </button>
          <button
            onClick={() => setFilterMode('app')}
            style={{
              padding: '0.35rem 0.75rem',
              fontSize: '0.75rem',
              fontWeight: 600,
              borderRadius: '4px',
              border: 'none',
              background: filterMode === 'app' ? 'rgba(168, 85, 247, 0.2)' : 'transparent',
              color: filterMode === 'app' ? '#a855f7' : 'var(--text-muted)',
              cursor: 'pointer',
            }}
          >
            App Libs ({report?.appPackages || 0})
          </button>
          <button
            onClick={() => setFilterMode('os')}
            style={{
              padding: '0.35rem 0.75rem',
              fontSize: '0.75rem',
              fontWeight: 600,
              borderRadius: '4px',
              border: 'none',
              background: filterMode === 'os' ? 'rgba(56, 189, 248, 0.2)' : 'transparent',
              color: filterMode === 'os' ? 'var(--accent)' : 'var(--text-muted)',
              cursor: 'pointer',
            }}
          >
            OS Distro ({report?.osPackages || 0})
          </button>
        </div>
      </div>

      {/* Packages Table */}
      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8125rem' }}>
          <thead>
            <tr style={{ background: 'var(--bg-surface)', borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
              <th style={{ padding: '0.75rem 1rem', width: '30px' }}></th>
              <th style={{ padding: '0.75rem 1rem' }}>Ecosystem</th>
              <th style={{ padding: '0.75rem 1rem' }}>Package Name</th>
              <th style={{ padding: '0.75rem 1rem' }}>Version</th>
              <th style={{ padding: '0.75rem 1rem' }}>Vulnerabilities (OSV)</th>
              <th style={{ padding: '0.75rem 1rem' }}>License</th>
              <th style={{ padding: '0.75rem 1rem' }}>Package URL (PURL)</th>
            </tr>
          </thead>
          <tbody>
            {filteredPackages.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
                  No matching components found.
                </td>
              </tr>
            ) : (
              filteredPackages.map((pkg, idx) => {
                const pkgKey = `${pkg.type}:${pkg.name}@${pkg.version}`;
                const isExpanded = expandedPkg === pkgKey;
                const hasVulns = pkg.vulnerabilities && pkg.vulnerabilities.length > 0;

                return (
                  <React.Fragment key={idx}>
                    <tr
                      onClick={() => hasVulns && setExpandedPkg(isExpanded ? null : pkgKey)}
                      style={{
                        borderBottom: '1px solid rgba(30,41,59,0.5)',
                        cursor: hasVulns ? 'pointer' : 'default',
                        background: hasVulns ? 'rgba(239, 68, 68, 0.04)' : undefined,
                      }}
                    >
                      <td style={{ padding: '0.65rem 0.5rem 0.65rem 1rem', color: 'var(--text-muted)' }}>
                        {hasVulns ? (
                          isExpanded ? <ChevronDown size={14} color="#f87171" /> : <ChevronRight size={14} color="#f87171" />
                        ) : null}
                      </td>
                      <td style={{ padding: '0.65rem 1rem' }}>
                        <span
                          style={{
                            padding: '0.15rem 0.45rem',
                            borderRadius: '4px',
                            fontSize: '0.7rem',
                            fontWeight: 700,
                            background: pkg.type.startsWith('os') ? 'rgba(56,189,248,0.15)' : 'rgba(168,85,247,0.15)',
                            color: pkg.type.startsWith('os') ? 'var(--accent)' : 'var(--purple)',
                          }}
                        >
                          {pkg.type}
                        </span>
                      </td>
                      <td style={{ padding: '0.65rem 1rem', fontWeight: 600 }}>{pkg.name}</td>
                      <td style={{ padding: '0.65rem 1rem', fontFamily: 'JetBrains Mono, monospace' }}>{pkg.version}</td>
                      <td style={{ padding: '0.65rem 1rem' }}>
                        {hasVulns ? (
                          <span
                            style={{
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: '0.3rem',
                              padding: '0.2rem 0.5rem',
                              borderRadius: '4px',
                              fontSize: '0.725rem',
                              fontWeight: 700,
                              background: 'rgba(239, 68, 68, 0.15)',
                              color: '#ef4444',
                              border: '1px solid rgba(239, 68, 68, 0.3)',
                            }}
                          >
                            <ShieldAlert size={12} /> {pkg.vulnerabilities!.length} Known CVE{pkg.vulnerabilities!.length > 1 ? 's' : ''}
                          </span>
                        ) : (
                          <span
                            style={{
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: '0.3rem',
                              padding: '0.2rem 0.45rem',
                              borderRadius: '4px',
                              fontSize: '0.725rem',
                              fontWeight: 600,
                              color: 'var(--emerald)',
                            }}
                          >
                            <ShieldCheck size={12} /> Clean
                          </span>
                        )}
                      </td>
                      <td style={{ padding: '0.65rem 1rem', color: 'var(--text-muted)' }}>{pkg.license || 'N/A'}</td>
                      <td style={{ padding: '0.65rem 1rem', fontFamily: 'JetBrains Mono, monospace', color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                        <code>{pkg.purl}</code>
                      </td>
                    </tr>

                    {/* Expanded CVE details drawer row */}
                    {isExpanded && hasVulns && (
                      <tr style={{ background: 'rgba(15, 23, 42, 0.7)' }}>
                        <td colSpan={7} style={{ padding: '1rem 1.5rem', borderBottom: '1px solid var(--border)' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                            <div style={{ fontSize: '0.8125rem', fontWeight: 700, color: '#f87171', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                              <ShieldAlert size={14} /> Security Advisories for {pkg.name}@{pkg.version}
                            </div>
                            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '0.75rem' }}>
                              {pkg.vulnerabilities!.map((v, vIdx) => (
                                <div
                                  key={vIdx}
                                  style={{
                                    background: 'var(--bg-card)',
                                    border: '1px solid rgba(239, 68, 68, 0.25)',
                                    borderRadius: '6px',
                                    padding: '0.75rem',
                                  }}
                                >
                                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                                      <span style={{ fontWeight: 700, color: 'white', fontFamily: 'JetBrains Mono, monospace', fontSize: '0.8rem' }}>
                                        {v.id}
                                      </span>
                                      {v.aliases && v.aliases.length > 0 && (
                                        <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>
                                          ({v.aliases.join(', ')})
                                        </span>
                                      )}
                                    </div>
                                    <span
                                      style={{
                                        fontSize: '0.7rem',
                                        fontWeight: 800,
                                        padding: '0.15rem 0.4rem',
                                        borderRadius: '3px',
                                        background: 'rgba(239, 68, 68, 0.2)',
                                        color: '#ef4444',
                                      }}
                                    >
                                      {v.severity || 'HIGH'}
                                    </span>
                                  </div>
                                  <p style={{ color: 'var(--text-muted)', fontSize: '0.775rem', lineHeight: 1.4, margin: '0 0 0.5rem 0' }}>
                                    {v.summary || 'Security vulnerability detected in this component.'}
                                  </p>
                                  {v.fixedIn && v.fixedIn.length > 0 && (
                                    <div style={{ fontSize: '0.725rem', color: 'var(--emerald)', fontWeight: 600 }}>
                                      Fixed in: {v.fixedIn.join(', ')}
                                    </div>
                                  )}
                                  <div style={{ marginTop: '0.5rem', textAlign: 'right' }}>
                                    <a
                                      href={`https://osv.dev/vulnerability/${v.id}`}
                                      target="_blank"
                                      rel="noreferrer"
                                      style={{
                                        fontSize: '0.7rem',
                                        color: 'var(--accent)',
                                        textDecoration: 'none',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '0.2rem',
                                      }}
                                    >
                                      View OSV details <ExternalLink size={10} />
                                    </a>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
