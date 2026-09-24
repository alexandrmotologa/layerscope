import React, { useState, useMemo } from 'react';
import {
  Search,
  Folder,
  File,
  AlertTriangle,
  X,
  Copy,
  Check,
  Binary,
  Layers,
  FileCode,
  Archive,
  Settings,
  Library,
} from 'lucide-react';
import { VFSNode, FilePreview } from '../types';
import { api } from '../api/client';

interface FileTreeProps {
  nodes: Record<string, VFSNode>;
  layerIndex?: number;
}

type CategoryFilter = 'all' | 'binaries' | 'caches' | 'config' | 'libraries' | 'wasted';
type SizeFilter = 'all' | '100k' | '1m' | '10m';

export const FileTree: React.FC<FileTreeProps> = ({ nodes, layerIndex = 0 }) => {
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const [sizeFilter, setSizeFilter] = useState<SizeFilter>('all');
  const [selectedFile, setSelectedFile] = useState<FilePreview | null>(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleFileClick = async (node: VFSNode) => {
    if (node.isDir) return;
    setLoadingPreview(true);
    setPreviewError(null);
    try {
      const preview = await api.getFilePreview(layerIndex, node.path);
      setSelectedFile(preview);
    } catch (err: any) {
      setPreviewError(err.message || 'Could not load file content');
      setSelectedFile({
        path: node.path,
        name: node.name,
        size: node.size,
        mode: '-rw-r--r--',
        modTime: '',
        isDir: false,
        isSymlink: false,
        layerIndex,
        changeType: node.changeType,
        isWasted: node.isWasted,
        wastedBytes: node.wastedBytes,
        wasteReason: node.wasteReason,
        isText: false,
        content: `Error loading content: ${err.message}`,
        lineCount: 1,
        truncated: false,
        mimeType: 'text/plain',
      });
    } finally {
      setLoadingPreview(false);
    }
  };

  const handleCopy = () => {
    if (selectedFile?.content) {
      navigator.clipboard.writeText(selectedFile.content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const isCategoryMatch = (p: string, cat: CategoryFilter): boolean => {
    if (cat === 'all') return true;
    const lower = p.toLowerCase();
    switch (cat) {
      case 'binaries':
        return (
          lower.startsWith('/bin') ||
          lower.startsWith('/sbin') ||
          lower.startsWith('/usr/bin') ||
          lower.startsWith('/usr/sbin') ||
          lower.endsWith('.so') ||
          lower.endsWith('.dylib') ||
          lower.endsWith('.a')
        );
      case 'caches':
        return (
          lower.startsWith('/tmp') ||
          lower.startsWith('/var/cache') ||
          lower.includes('.cache') ||
          lower.includes('.npm') ||
          lower.includes('var/lib/apt/lists')
        );
      case 'config':
        return (
          lower.startsWith('/etc') ||
          lower.endsWith('.conf') ||
          lower.endsWith('.json') ||
          lower.endsWith('.yaml') ||
          lower.endsWith('.yml') ||
          lower.endsWith('.ini') ||
          lower.endsWith('.env') ||
          lower.includes('.env.')
        );
      case 'libraries':
        return (
          lower.startsWith('/lib') ||
          lower.startsWith('/usr/lib') ||
          lower.includes('node_modules') ||
          lower.includes('site-packages')
        );
      case 'wasted':
        return true; // handled separately
      default:
        return true;
    }
  };

  const isSizeMatch = (size: number, sf: SizeFilter): boolean => {
    switch (sf) {
      case '100k':
        return size >= 100 * 1024;
      case '1m':
        return size >= 1024 * 1024;
      case '10m':
        return size >= 10 * 1024 * 1024;
      default:
        return true;
    }
  };

  const filteredPaths = useMemo(() => {
    const q = search.toLowerCase().trim();
    return Object.keys(nodes)
      .filter((p) => {
        const node = nodes[p];
        if (q && !p.toLowerCase().includes(q)) return false;
        if (category === 'wasted' && !node.isWasted && !node.isDir) return false;
        if (!isCategoryMatch(p, category) && !node.isDir) return false;
        if (!isSizeMatch(node.size, sizeFilter) && !node.isDir) return false;
        return true;
      })
      .sort();
  }, [nodes, search, category, sizeFilter]);

  return (
    <div style={{ flex: 1, display: 'flex', position: 'relative', overflow: 'hidden', background: 'var(--bg-dark)' }}>
      {/* Main File Tree Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {/* Filter Controls Bar */}
        <div
          style={{
            padding: '0.75rem 1rem',
            borderBottom: '1px solid var(--border)',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.6rem',
            background: 'var(--bg-surface)',
          }}
        >
          {/* Search bar + size threshold dropdown */}
          <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
            <div style={{ position: 'relative', flex: 1, display: 'flex', alignItems: 'center' }}>
              <Search size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.75rem' }} />
              <input
                type="text"
                placeholder="Search files by path (e.g. /etc, /app, .env, .json)..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                style={{
                  width: '100%',
                  background: 'var(--bg-card)',
                  border: '1px solid var(--border)',
                  padding: '0.45rem 0.875rem 0.45rem 2.25rem',
                  borderRadius: '6px',
                  color: 'var(--text)',
                  fontSize: '0.875rem',
                  outline: 'none',
                  fontFamily: 'Inter, sans-serif',
                }}
              />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
              <span>Size:</span>
              <select
                value={sizeFilter}
                onChange={(e) => setSizeFilter(e.target.value as SizeFilter)}
                style={{
                  background: 'var(--bg-card)',
                  border: '1px solid var(--border)',
                  color: 'var(--text)',
                  padding: '0.4rem 0.6rem',
                  borderRadius: '6px',
                  fontSize: '0.8125rem',
                  outline: 'none',
                  cursor: 'pointer',
                }}
              >
                <option value="all">All sizes</option>
                <option value="100k">&gt; 100 KB</option>
                <option value="1m">&gt; 1 MB</option>
                <option value="10m">&gt; 10 MB</option>
              </select>
            </div>
          </div>

          {/* Category Tabs */}
          <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', alignItems: 'center' }}>
            {(
              [
                { id: 'all', label: 'All Files', icon: Layers },
                { id: 'binaries', label: 'Binaries', icon: Binary },
                { id: 'caches', label: 'Caches', icon: Archive },
                { id: 'config', label: 'Config', icon: Settings },
                { id: 'libraries', label: 'Libraries', icon: Library },
                { id: 'wasted', label: 'Wasted Only', icon: AlertTriangle },
              ] as const
            ).map((cat) => {
              const Icon = cat.icon;
              const active = category === cat.id;
              return (
                <button
                  key={cat.id}
                  onClick={() => setCategory(cat.id)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.35rem',
                    padding: '0.3rem 0.65rem',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    borderRadius: '5px',
                    border: active ? '1px solid var(--accent)' : '1px solid var(--border)',
                    background: active ? 'rgba(56, 189, 248, 0.15)' : 'var(--bg-card)',
                    color: active ? 'var(--accent)' : 'var(--text-muted)',
                    cursor: 'pointer',
                    transition: 'all 0.15s ease',
                  }}
                >
                  <Icon size={12} />
                  <span>{cat.label}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Nodes list */}
        <div
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '0.5rem 1rem',
            fontFamily: 'JetBrains Mono, monospace',
            fontSize: '0.8125rem',
          }}
        >
          {filteredPaths.length === 0 ? (
            <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
              No files match the active filters.
            </div>
          ) : (
            filteredPaths.slice(0, 1500).map((path) => {
              const node = nodes[path];
              const sizeStr = node.isDir
                ? ''
                : node.size > 1024 * 1024
                ? `${(node.size / (1024 * 1024)).toFixed(1)} MB`
                : `${(node.size / 1024).toFixed(0)} KB`;

              const isSelected = selectedFile?.path === node.path;

              return (
                <div
                  key={path}
                  onClick={() => handleFileClick(node)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    padding: '0.35rem 0.5rem',
                    borderRadius: '4px',
                    gap: '0.5rem',
                    cursor: node.isDir ? 'default' : 'pointer',
                    background: isSelected
                      ? 'rgba(56, 189, 248, 0.12)'
                      : node.isWasted
                      ? 'rgba(168, 85, 247, 0.05)'
                      : undefined,
                    transition: 'background 0.1s ease',
                  }}
                >
                  <span style={{ color: 'var(--text-muted)' }}>
                    {node.isDir ? <Folder size={14} color="#60a5fa" /> : <File size={14} />}
                  </span>

                  <span
                    style={{
                      flex: 1,
                      color: node.isWasted ? 'var(--purple)' : 'var(--text)',
                      textDecoration: node.changeType === 'deleted' ? 'line-through' : undefined,
                    }}
                  >
                    {node.path}
                  </span>

                  {sizeStr && (
                    <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>{sizeStr}</span>
                  )}

                  {node.changeType === 'added' && (
                    <span
                      style={{
                        background: 'rgba(34,197,94,0.15)',
                        color: 'var(--green)',
                        padding: '0.1rem 0.4rem',
                        borderRadius: '4px',
                        fontSize: '0.7rem',
                        fontWeight: 700,
                      }}
                    >
                      + Added
                    </span>
                  )}
                  {node.changeType === 'modified' && (
                    <span
                      style={{
                        background: 'rgba(234,179,8,0.15)',
                        color: 'var(--yellow)',
                        padding: '0.1rem 0.4rem',
                        borderRadius: '4px',
                        fontSize: '0.7rem',
                        fontWeight: 700,
                      }}
                    >
                      ~ Modified
                    </span>
                  )}
                  {node.changeType === 'deleted' && (
                    <span
                      style={{
                        background: 'rgba(239,68,68,0.15)',
                        color: 'var(--red)',
                        padding: '0.1rem 0.4rem',
                        borderRadius: '4px',
                        fontSize: '0.7rem',
                        fontWeight: 700,
                      }}
                    >
                      - Deleted
                    </span>
                  )}

                  {node.isWasted && (
                    <span
                      title={node.wasteReason || 'Wasted space'}
                      style={{
                        background: 'rgba(168,85,247,0.2)',
                        color: 'var(--purple)',
                        padding: '0.1rem 0.4rem',
                        borderRadius: '4px',
                        fontSize: '0.7rem',
                        fontWeight: 700,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.25rem',
                      }}
                    >
                      <AlertTriangle size={11} />
                      Wasted ({(node.wastedBytes / 1024).toFixed(0)}KB)
                    </span>
                  )}
                </div>
              );
            })
          )}
        </div>
      </div>

      {/* Slide-out File Inspection & Preview Drawer */}
      {selectedFile && (
        <div
          style={{
            width: '460px',
            borderLeft: '1px solid var(--border)',
            background: 'var(--bg-surface)',
            display: 'flex',
            flexDirection: 'column',
            boxShadow: '-4px 0 24px rgba(0, 0, 0, 0.4)',
            zIndex: 10,
          }}
        >
          {/* Drawer Header */}
          <div
            style={{
              padding: '1rem',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              background: 'var(--bg-card)',
            }}
          >
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.25rem' }}>
                <FileCode size={16} color="var(--accent)" />
                <h4
                  style={{
                    fontSize: '0.925rem',
                    fontWeight: 700,
                    margin: 0,
                    color: 'white',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                  title={selectedFile.path}
                >
                  {selectedFile.name}
                </h4>
              </div>
              <div
                style={{
                  fontFamily: 'JetBrains Mono, monospace',
                  fontSize: '0.725rem',
                  color: 'var(--text-muted)',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {selectedFile.path}
              </div>
            </div>

            <button
              onClick={() => setSelectedFile(null)}
              style={{
                background: 'transparent',
                border: 'none',
                color: 'var(--text-muted)',
                cursor: 'pointer',
                padding: '0.25rem',
              }}
            >
              <X size={18} />
            </button>
          </div>

          {/* File Metadata Badges */}
          <div
            style={{
              padding: '0.75rem 1rem',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              gap: '0.5rem',
              flexWrap: 'wrap',
              fontSize: '0.75rem',
              background: 'rgba(15, 23, 42, 0.4)',
            }}
          >
            <span
              style={{
                padding: '0.2rem 0.5rem',
                borderRadius: '4px',
                background: 'rgba(255,255,255,0.06)',
                color: 'var(--text)',
                fontFamily: 'JetBrains Mono, monospace',
              }}
            >
              {(selectedFile.size / 1024).toFixed(1)} KB ({selectedFile.size} bytes)
            </span>
            <span
              style={{
                padding: '0.2rem 0.5rem',
                borderRadius: '4px',
                background: 'rgba(255,255,255,0.06)',
                color: 'var(--text-muted)',
                fontFamily: 'JetBrains Mono, monospace',
              }}
            >
              Mode: {selectedFile.mode}
            </span>
            <span
              style={{
                padding: '0.2rem 0.5rem',
                borderRadius: '4px',
                background: 'rgba(255,255,255,0.06)',
                color: 'var(--text-muted)',
              }}
            >
              {selectedFile.mimeType}
            </span>
            {selectedFile.isWasted && (
              <span
                style={{
                  padding: '0.2rem 0.5rem',
                  borderRadius: '4px',
                  background: 'rgba(168, 85, 247, 0.2)',
                  color: 'var(--purple)',
                  fontWeight: 700,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.25rem',
                }}
              >
                <AlertTriangle size={11} /> Wasted Space
              </span>
            )}
          </div>

          {/* Preview Toolbar */}
          <div
            style={{
              padding: '0.5rem 1rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              borderBottom: '1px solid var(--border)',
              fontSize: '0.75rem',
              color: 'var(--text-muted)',
            }}
          >
            <span>{selectedFile.isText ? `${selectedFile.lineCount} lines` : 'Binary content'}</span>
            {selectedFile.isText && selectedFile.content && (
              <button
                onClick={handleCopy}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.3rem',
                  background: 'var(--bg-card)',
                  border: '1px solid var(--border)',
                  color: 'var(--text)',
                  padding: '0.25rem 0.6rem',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  fontSize: '0.725rem',
                }}
              >
                {copied ? <Check size={12} color="var(--emerald)" /> : <Copy size={12} />}
                <span>{copied ? 'Copied!' : 'Copy Content'}</span>
              </button>
            )}
          </div>

          {/* Content Viewer */}
          <div
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: '1rem',
              background: '#090d16',
              fontFamily: 'JetBrains Mono, monospace',
              fontSize: '0.775rem',
              lineHeight: 1.5,
              color: '#e2e8f0',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-all',
            }}
          >
            {loadingPreview ? (
              <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
                Loading file preview...
              </div>
            ) : previewError ? (
              <div style={{ color: 'var(--red)', padding: '1rem' }}>
                {previewError}
              </div>
            ) : selectedFile.isText && selectedFile.content ? (
              selectedFile.content
            ) : (
              <div style={{ textAlign: 'center', padding: '3rem 1rem', color: 'var(--text-muted)' }}>
                <Binary size={36} color="var(--text-muted)" style={{ marginBottom: '0.75rem' }} />
                <p style={{ margin: '0 0 0.5rem 0', fontWeight: 600 }}>Binary File Preview Unavailable</p>
                <p style={{ margin: 0, fontSize: '0.725rem' }}>
                  This file contains raw binary data or exceeds the 256KB inline preview limit.
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
