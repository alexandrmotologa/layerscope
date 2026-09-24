import {
  ImageData,
  LayerTreeResponse,
  SecurityReport,
  SBOMReport,
  AdvisorReport,
  DiffReport,
  FilePreview,
  DockerfileOptimization,
  ImagePreset,
} from '../types';

export const api = {
  async getImage(): Promise<ImageData> {
    const res = await fetch('/api/image');
    if (!res.ok) throw new Error('Failed to load image metadata');
    return res.json();
  },

  async getLayerTree(layerIndex: number, wastedOnly = false, prefix = ''): Promise<LayerTreeResponse> {
    const params = new URLSearchParams();
    if (wastedOnly) params.set('wastedOnly', 'true');
    if (prefix) params.set('prefix', prefix);

    const res = await fetch(`/api/layers/${layerIndex}/tree?${params.toString()}`);
    if (!res.ok) throw new Error('Failed to load layer tree');
    return res.json();
  },

  async getFilePreview(layerIndex: number, path: string): Promise<FilePreview> {
    const params = new URLSearchParams({ path });
    const res = await fetch(`/api/layers/${layerIndex}/file?${params.toString()}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Failed to inspect file content');
    }
    return res.json();
  },

  async getSecurity(): Promise<SecurityReport> {
    const res = await fetch('/api/security');
    if (!res.ok) throw new Error('Failed to load security report');
    return res.json();
  },

  async getSBOM(): Promise<SBOMReport> {
    const res = await fetch('/api/sbom');
    if (!res.ok) throw new Error('Failed to load SBOM');
    return res.json();
  },

  async getAdvisor(): Promise<AdvisorReport> {
    const res = await fetch('/api/advisor');
    if (!res.ok) throw new Error('Failed to load advisor report');
    return res.json();
  },

  async getDockerfileOptimization(): Promise<DockerfileOptimization> {
    const res = await fetch('/api/advisor/dockerfile');
    if (!res.ok) throw new Error('Failed to load Dockerfile optimization');
    return res.json();
  },

  async getPresets(): Promise<ImagePreset[]> {
    const res = await fetch('/api/presets');
    if (!res.ok) throw new Error('Failed to load presets');
    return res.json();
  },

  async analyzeImage(target: string, forceRemote = false, architecture = 'amd64', demo = false): Promise<any> {
    const res = await fetch('/api/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target, forceRemote, architecture, demo }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Failed to analyze image');
    }
    return res.json();
  },

  async diff(imageA: string, imageB: string, archA = 'amd64', archB = 'amd64'): Promise<DiffReport> {
    const res = await fetch('/api/diff', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ imageA, imageB, archA, archB }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to compare images');
    }
    return res.json();
  },
};
