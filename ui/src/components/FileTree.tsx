import React, { useState, useMemo } from 'react';
import { Search, Folder, File, AlertTriangle } from 'lucide-react';
import { VFSNode } from '../types';

interface FileTreeProps {
  nodes: Record<string, VFSNode>;
}

export const FileTree: React.FC<FileTreeProps> = ({ nodes }) => {
  const [search, setSearch] = useState('');
  const [wastedOnly, setWastedOnly] = useState(false);

  const filteredPaths = useMemo(() => {
    const q = search.toLowerCase().trim();
    return Object.keys(nodes)
      .filter((p) => {
        const node = nodes[p];
        if (q && !p.toLowerCase().includes(q)) return false;
        if (wastedOnly && !node.isWasted && !node.isDir) return false;
        return true;
      })
      .sort();
  }, [nodes, search, wastedOnly]);

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', background: 'var(--bg-dark)' }}>
      {/* Top Filter Bar */}
      <div style={{
        padding: '0.75rem 1rem',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        gap: '0.75rem',
        alignItems: 'center',
        background: 'var(--bg-surface)',
      }}>
        <div style={{
          position: 'relative',
          flex: 1,
          display: 'flex',
          alignItems: 'center',
        }}>
          <Search size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.75rem' }} />
          <input
            type="text"
            placeholder="Search files by path (e.g. /etc, /app, .env, .js)..."
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

        <label style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          fontSize: '0.8125rem',
          cursor: 'pointer',
          color: 'var(--purple)',
          fontWeight: 600,
        }}>
          <input
            type="checkbox"
            checked={wastedOnly}
            onChange={(e) => setWastedOnly(e.target.checked)}
          />
          <span>Only Wasted Files</span>
        </label>
      </div>

      {/* Nodes list */}
      <div style={{
        flex: 1,
        overflowY: 'auto',
        padding: '0.5rem 1rem',
        fontFamily: 'JetBrains Mono, monospace',
        fontSize: '0.8125rem',
      }}>
        {filteredPaths.length === 0 ? (
          <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No files match the selected filter.
          </div>
        ) : (
          filteredPaths.slice(0, 1500).map((path) => {
            const node = nodes[path];
            const sizeStr = node.isDir
              ? ''
              : node.size > 1024 * 1024
              ? `${(node.size / (1024 * 1024)).toFixed(1)} MB`
              : `${(node.size / 1024).toFixed(0)} KB`;

            return (
              <div
                key={path}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  padding: '0.35rem 0.5rem',
                  borderRadius: '4px',
                  gap: '0.5rem',
                }}
              >
                <span style={{ color: 'var(--text-muted)' }}>
                  {node.isDir ? <Folder size={14} color="#60a5fa" /> : <File size={14} />}
                </span>

                <span style={{ flex: 1, color: node.isWasted ? 'var(--purple)' : 'var(--text)' }}>
                  {node.path}
                </span>

                {sizeStr && (
                  <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                    {sizeStr}
                  </span>
                )}

                {node.changeType === 'added' && (
                  <span style={{ background: 'rgba(34,197,94,0.15)', color: 'var(--green)', padding: '0.1rem 0.4rem', borderRadius: '4px', fontSize: '0.7rem', fontWeight: 700 }}>
                    + Added
                  </span>
                )}
                {node.changeType === 'modified' && (
                  <span style={{ background: 'rgba(234,179,8,0.15)', color: 'var(--yellow)', padding: '0.1rem 0.4rem', borderRadius: '4px', fontSize: '0.7rem', fontWeight: 700 }}>
                    ~ Modified
                  </span>
                )}
                {node.changeType === 'deleted' && (
                  <span style={{ background: 'rgba(239,68,68,0.15)', color: 'var(--red)', padding: '0.1rem 0.4rem', borderRadius: '4px', fontSize: '0.7rem', fontWeight: 700 }}>
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
  );
};
