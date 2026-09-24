import React from 'react';
import { LayerSummary } from '../types';

interface LayerTimelineProps {
  layers: LayerSummary[];
  activeLayerIndex: number;
  onSelectLayer: (index: number) => void;
}

export const LayerTimeline: React.FC<LayerTimelineProps> = ({
  layers,
  activeLayerIndex,
  onSelectLayer,
}) => {
  return (
    <div style={{
      width: '380px',
      borderRight: '1px solid var(--border)',
      background: 'var(--bg-card)',
      display: 'flex',
      flexDirection: 'column',
      height: '100%',
    }}>
      <div style={{
        padding: '0.75rem 1rem',
        borderBottom: '1px solid var(--border)',
        fontSize: '0.8125rem',
        fontWeight: 600,
        color: 'var(--text-muted)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <span>IMAGE LAYERS</span>
        <span>{layers.length} layers</span>
      </div>

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {layers.map((layer) => {
          const isSelected = layer.index === activeLayerIndex;
          const sizeMB = (layer.size / (1024 * 1024)).toFixed(1);
          const wastedKB = (layer.wastedBytes / 1024).toFixed(0);

          return (
            <div
              key={layer.index}
              onClick={() => onSelectLayer(layer.index)}
              style={{
                padding: '0.875rem 1rem',
                borderBottom: '1px solid rgba(30, 41, 59, 0.5)',
                cursor: 'pointer',
                background: isSelected ? 'var(--bg-hover)' : 'transparent',
                borderLeft: isSelected ? '3px solid var(--accent)' : '3px solid transparent',
                transition: 'background 0.12s ease',
              }}
            >
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginBottom: '0.25rem',
                fontSize: '0.8125rem',
              }}>
                <strong style={{ color: isSelected ? 'var(--accent)' : 'var(--text)' }}>
                  Layer {layer.index}
                </strong>
                <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
                  {layer.wastedBytes > 0 && (
                    <span style={{
                      background: 'rgba(168, 85, 247, 0.2)',
                      color: 'var(--purple)',
                      fontSize: '0.7rem',
                      padding: '0.1rem 0.4rem',
                      borderRadius: '4px',
                      fontWeight: 600,
                    }}>
                      {wastedKB}KB Wasted
                    </span>
                  )}
                  <span style={{ fontFamily: 'JetBrains Mono, monospace', color: 'var(--text-muted)' }}>
                    {sizeMB} MB
                  </span>
                </div>
              </div>

              <div style={{
                fontFamily: 'JetBrains Mono, monospace',
                fontSize: '0.75rem',
                color: '#cbd5e1',
                wordBreak: 'break-all',
                lineHeight: 1.4,
              }}>
                {layer.command || 'Base system layer'}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
