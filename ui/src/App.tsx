import React, { useState, useEffect } from 'react';
import { Layers, GitCompare, ShieldAlert, Package, Sparkles } from 'lucide-react';
import { Header } from './components/Header';
import { LayerTimeline } from './components/LayerTimeline';
import { FileTree } from './components/FileTree';
import { DiffWorkspace } from './components/DiffWorkspace';
import { SecretAuditor } from './components/SecretAuditor';
import { SbomViewer } from './components/SbomViewer';
import { AdvisorPanel } from './components/AdvisorPanel';
import { api } from './api/client';
import { ImageData, VFSNode, SecurityReport, SBOMReport, AdvisorReport } from './types';

type Tab = 'inspector' | 'diff' | 'security' | 'sbom' | 'advisor';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<Tab>('inspector');
  const [imageData, setImageData] = useState<ImageData | null>(null);
  const [activeLayerIndex, setActiveLayerIndex] = useState(0);
  const [layerNodes, setLayerNodes] = useState<Record<string, VFSNode>>({});
  const [securityData, setSecurityData] = useState<SecurityReport | null>(null);
  const [sbomData, setSbomData] = useState<SBOMReport | null>(null);
  const [advisorData, setAdvisorData] = useState<AdvisorReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function init() {
      try {
        const img = await api.getImage();
        setImageData(img);
        if (img.layers.length > 0) {
          const tree = await api.getLayerTree(0);
          setLayerNodes(tree.nodes);
        }
        // Prefetch other modules in background
        api.getSecurity().then(setSecurityData).catch(console.error);
        api.getSBOM().then(setSbomData).catch(console.error);
        api.getAdvisor().then(setAdvisorData).catch(console.error);
      } catch (e) {
        console.error('Failed to initialize studio:', e);
      } finally {
        setLoading(false);
      }
    }
    init();
  }, []);

  const refreshAll = async () => {
    setLoading(true);
    try {
      const img = await api.getImage();
      setImageData(img);
      setActiveLayerIndex(0);
      if (img.layers.length > 0) {
        const tree = await api.getLayerTree(0);
        setLayerNodes(tree.nodes);
      }
      const [sec, sbom, adv] = await Promise.all([
        api.getSecurity().catch(() => null),
        api.getSBOM().catch(() => null),
        api.getAdvisor().catch(() => null),
      ]);
      setSecurityData(sec);
      setSbomData(sbom);
      setAdvisorData(adv);
    } catch (e) {
      console.error('Failed to reload studio:', e);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectLayer = async (index: number) => {
    setActiveLayerIndex(index);
    try {
      const tree = await api.getLayerTree(index);
      setLayerNodes(tree.nodes);
    } catch (e) {
      console.error('Failed to load layer tree:', e);
    }
  };

  if (loading) {
    return (
      <div style={{ height: '100vh', display: 'grid', placeItems: 'center', background: 'var(--bg-dark)' }}>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
          <div style={{ width: '40px', height: '40px', border: '3px solid var(--border)', borderTopColor: 'var(--accent)', borderRadius: '50%', animation: 'spin 1s linear infinite' }} />
          <div style={{ fontSize: '0.9rem', color: 'var(--text-muted)' }}>Loading LayerScope Studio...</div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', overflow: 'hidden' }}>
      <Header imageData={imageData} securityData={securityData} onImageChanged={refreshAll} />

      {/* Tab Navigation */}
      <nav style={{
        background: '#0b1120',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        padding: '0 1.5rem',
        gap: '0.25rem',
      }}>
        <button
          onClick={() => setActiveTab('inspector')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'inspector' ? 'var(--accent)' : 'var(--text-muted)',
            borderBottom: activeTab === 'inspector' ? '2px solid var(--accent)' : '2px solid transparent',
            padding: '0.75rem 1rem',
            fontWeight: 600,
            fontSize: '0.875rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <Layers size={16} /> Layer Inspector
        </button>

        <button
          onClick={() => setActiveTab('diff')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'diff' ? 'var(--accent)' : 'var(--text-muted)',
            borderBottom: activeTab === 'diff' ? '2px solid var(--accent)' : '2px solid transparent',
            padding: '0.75rem 1rem',
            fontWeight: 600,
            fontSize: '0.875rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <GitCompare size={16} /> Side-by-Side Diff
        </button>

        <button
          onClick={() => setActiveTab('security')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'security' ? 'var(--accent)' : 'var(--text-muted)',
            borderBottom: activeTab === 'security' ? '2px solid var(--accent)' : '2px solid transparent',
            padding: '0.75rem 1rem',
            fontWeight: 600,
            fontSize: '0.875rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <ShieldAlert size={16} /> Secret Audit
          {securityData && securityData.totalFindings > 0 && (
            <span style={{
              background: '#dc2626',
              color: 'white',
              fontSize: '0.7rem',
              padding: '0.1rem 0.45rem',
              borderRadius: '9999px',
              fontWeight: 700,
            }}>
              {securityData.totalFindings}
            </span>
          )}
        </button>

        <button
          onClick={() => setActiveTab('sbom')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'sbom' ? 'var(--accent)' : 'var(--text-muted)',
            borderBottom: activeTab === 'sbom' ? '2px solid var(--accent)' : '2px solid transparent',
            padding: '0.75rem 1rem',
            fontWeight: 600,
            fontSize: '0.875rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <Package size={16} /> SBOM Packages
        </button>

        <button
          onClick={() => setActiveTab('advisor')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'advisor' ? 'var(--accent)' : 'var(--text-muted)',
            borderBottom: activeTab === 'advisor' ? '2px solid var(--accent)' : '2px solid transparent',
            padding: '0.75rem 1rem',
            fontWeight: 600,
            fontSize: '0.875rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <Sparkles size={16} /> Optimization Advisor
        </button>
      </nav>

      {/* Main Content Areas */}
      <main style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {activeTab === 'inspector' && (
          <div style={{ flex: 1, display: 'flex', width: '100%', height: '100%' }}>
            <LayerTimeline
              layers={imageData?.layers || []}
              activeLayerIndex={activeLayerIndex}
              onSelectLayer={handleSelectLayer}
            />
            <FileTree
              nodes={layerNodes}
              layerIndex={activeLayerIndex}
            />
          </div>
        )}

        {activeTab === 'diff' && (
          <DiffWorkspace currentImageName={imageData?.reference.original || ''} />
        )}

        {activeTab === 'security' && (
          <SecretAuditor report={securityData} />
        )}

        {activeTab === 'sbom' && (
          <SbomViewer report={sbomData} />
        )}

        {activeTab === 'advisor' && (
          <AdvisorPanel report={advisorData} />
        )}
      </main>
    </div>
  );
};
