package server

import (
	"fmt"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// GetFallbackStudioHTML renders a complete, responsive interactive dark-themed visual studio.
func GetFallbackStudioHTML(
	img *oci.ImageAnalysis,
	adv *advisor.AdvisorReport,
	waste *vfs.WastedSpaceSummary,
	sec *security.SecurityAuditReport,
) string {
	imageName := "No Image Loaded"
	totalSizeMB := 0.0
	effScore := 100.0
	grade := "A+"
	wastedMB := 0.0
	arch := "linux/amd64"
	secCount := 0
	intermediateCount := 0

	if img != nil {
		imageName = img.Reference.Original
		totalSizeMB = float64(img.TotalSizeBytes) / (1024 * 1024)
		arch = img.Config.Architecture
		if img.Config.OS != "" {
			arch = img.Config.OS + "/" + arch
		}
	}
	if adv != nil {
		effScore = adv.EfficiencyScore
		grade = adv.Grade
	}
	if waste != nil {
		wastedMB = float64(waste.TotalWastedBytes) / (1024 * 1024)
	}
	if sec != nil {
		secCount = sec.TotalFindings
		intermediateCount = sec.IntermediateLeaks
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LayerScope Studio - %s</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-dark: #080b11;
      --bg-card: #0f172a;
      --bg-hover: #1e293b;
      --border: #1e293b;
      --border-focus: #3b82f6;
      --text: #f1f5f9;
      --text-muted: #94a3b8;
      --accent: #38bdf8;
      --green: #22c55e;
      --yellow: #eab308;
      --red: #ef4444;
      --purple: #a855f7;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: 'Inter', -apple-system, sans-serif;
      background: var(--bg-dark);
      color: var(--text);
      display: flex;
      flex-direction: column;
      height: 100vh;
      overflow: hidden;
    }
    header {
      background: rgba(15, 23, 42, 0.85);
      backdrop-filter: blur(12px);
      border-bottom: 1px solid var(--border);
      padding: 0.75rem 1.5rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
      z-index: 10;
    }
    .brand { display: flex; align-items: center; gap: 0.75rem; font-weight: 800; font-size: 1.25rem; letter-spacing: -0.025em; }
    .brand-icon { width: 32px; height: 32px; background: linear-gradient(135deg, #38bdf8, #6366f1); border-radius: 8px; display: grid; place-items: center; font-size: 1.1rem; }
    .meta-pills { display: flex; gap: 0.5rem; align-items: center; font-size: 0.8125rem; }
    .pill {
      background: var(--bg-card);
      border: 1px solid var(--border);
      padding: 0.35rem 0.75rem;
      border-radius: 9999px;
      display: flex;
      align-items: center;
      gap: 0.375rem;
      font-family: 'JetBrains Mono', monospace;
    }
    .pill.grade { background: #064e3b; border-color: #059669; color: #34d399; font-weight: 700; }
    .pill.warn { background: #451a03; border-color: #d97706; color: #fbbf24; }
    .pill.danger { background: #450a0a; border-color: #dc2626; color: #f87171; }

    nav.tabs {
      background: #0b1120;
      border-bottom: 1px solid var(--border);
      display: flex;
      padding: 0 1.5rem;
      gap: 0.25rem;
    }
    .tab-btn {
      background: none;
      border: none;
      color: var(--text-muted);
      padding: 0.75rem 1rem;
      font-weight: 600;
      font-size: 0.875rem;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 0.5rem;
      border-bottom: 2px solid transparent;
      transition: all 0.15s ease;
    }
    .tab-btn:hover { color: var(--text); }
    .tab-btn.active { color: var(--accent); border-bottom-color: var(--accent); }
    .tab-badge { background: var(--bg-hover); padding: 0.15rem 0.45rem; border-radius: 9999px; font-size: 0.75rem; }

    main { flex: 1; display: flex; overflow: hidden; position: relative; }
    .view-panel { display: none; width: 100%%; height: 100%%; overflow: hidden; }
    .view-panel.active { display: flex; }

    /* Split Pane for Layer Inspector */
    .sidebar-layers {
      width: 380px;
      border-right: 1px solid var(--border);
      background: #090e1a;
      display: flex;
      flex-direction: column;
    }
    .pane-header {
      padding: 0.75rem 1rem;
      border-bottom: 1px solid var(--border);
      font-size: 0.8125rem;
      font-weight: 600;
      color: var(--text-muted);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .layer-list { flex: 1; overflow-y: auto; }
    .layer-item {
      padding: 0.875rem 1rem;
      border-bottom: 1px solid rgba(30, 41, 59, 0.5);
      cursor: pointer;
      transition: background 0.1s;
    }
    .layer-item:hover { background: var(--bg-hover); }
    .layer-item.selected { background: #1e293b; border-left: 3px solid var(--accent); }
    .layer-item-top { display: flex; justify-content: space-between; margin-bottom: 0.25rem; font-size: 0.8125rem; }
    .layer-cmd { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: #cbd5e1; word-break: break-all; }
    
    .center-tree { flex: 1; display: flex; flex-direction: column; background: var(--bg-dark); }
    .tree-filter-bar {
      padding: 0.75rem 1rem;
      border-bottom: 1px solid var(--border);
      display: flex;
      gap: 0.75rem;
      align-items: center;
    }
    .search-input {
      flex: 1;
      background: var(--bg-card);
      border: 1px solid var(--border);
      padding: 0.45rem 0.875rem;
      border-radius: 6px;
      color: white;
      font-size: 0.875rem;
      outline: none;
    }
    .search-input:focus { border-color: var(--accent); }
    .tree-scroll { flex: 1; overflow-y: auto; padding: 0.5rem 1rem; font-family: 'JetBrains Mono', monospace; font-size: 0.8125rem; }
    .tree-node-row {
      display: flex;
      align-items: center;
      padding: 0.35rem 0.5rem;
      border-radius: 4px;
      cursor: pointer;
    }
    .tree-node-row:hover { background: var(--bg-hover); }
    .node-icon { width: 20px; color: var(--text-muted); }
    .node-path { flex: 1; }
    .node-size { color: var(--text-muted); font-size: 0.75rem; margin-right: 0.75rem; }
    .tag { font-size: 0.7rem; padding: 0.15rem 0.45rem; border-radius: 4px; font-weight: 600; text-transform: uppercase; }
    .tag.added { background: rgba(34, 197, 94, 0.2); color: var(--green); }
    .tag.modified { background: rgba(234, 179, 8, 0.2); color: var(--yellow); }
    .tag.deleted { background: rgba(239, 68, 68, 0.2); color: var(--red); }
    .tag.wasted { background: rgba(168, 85, 247, 0.2); color: var(--purple); }

    /* Security Tab */
    .card-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1rem; padding: 1.5rem; }
    .stat-card { background: var(--bg-card); border: 1px solid var(--border); padding: 1.25rem; border-radius: 8px; }
    .stat-title { font-size: 0.8125rem; color: var(--text-muted); margin-bottom: 0.35rem; }
    .stat-value { font-size: 1.75rem; font-weight: 800; }
    .content-table-wrapper { padding: 0 1.5rem 1.5rem 1.5rem; flex: 1; overflow-y: auto; }
    table.data-table { width: 100%%; border-collapse: collapse; text-align: left; }
    table.data-table th { background: var(--bg-card); padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); font-size: 0.8125rem; color: var(--text-muted); }
    table.data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid rgba(30, 41, 59, 0.5); font-size: 0.875rem; }
    .btn { background: #2563eb; color: white; border: none; padding: 0.45rem 1rem; border-radius: 6px; font-weight: 600; cursor: pointer; }
    .btn:hover { background: #1d4ed8; }
    pre { background: #050811; padding: 1rem; border-radius: 6px; font-family: 'JetBrains Mono', monospace; font-size: 0.8125rem; overflow-x: auto; color: #38bdf8; margin-top: 0.75rem; }
  </style>
</head>
<body>

<header>
  <div class="brand">
    <div class="brand-icon">⬡</div>
    <div>LayerScope <span style="font-size:0.75rem; color:var(--text-muted); font-weight:normal;">Studio</span></div>
  </div>
  <div class="meta-pills">
    <div class="pill"><span>Image:</span> <strong>%s</strong></div>
    <div class="pill"><span>Arch:</span> <strong>%s</strong></div>
    <div class="pill"><span>Size:</span> <strong>%.1f MB</strong></div>
    <div class="pill grade"><span>Efficiency:</span> <strong>%.1f%% (%s)</strong></div>
    %s
  </div>
</header>

<nav class="tabs">
  <button class="tab-btn active" onclick="switchTab('inspector')">Layer Inspector</button>
  <button class="tab-btn" onclick="switchTab('diff')">Side-by-Side Diff</button>
  <button class="tab-btn" onclick="switchTab('security')">
    Secret Audit
    <span class="tab-badge" style="background:#dc2626; color:white;">%d</span>
  </button>
  <button class="tab-btn" onclick="switchTab('sbom')">SBOM Packages</button>
  <button class="tab-btn" onclick="switchTab('advisor')">Optimization Advisor</button>
</nav>

<main>
  <!-- 1. Layer Inspector -->
  <div id="view-inspector" class="view-panel active">
    <div class="sidebar-layers">
      <div class="pane-header">
        <span>IMAGE LAYERS</span>
        <span id="layer-count">0 layers</span>
      </div>
      <div class="layer-list" id="layer-list-container">
        <!-- Rendered by JS -->
      </div>
    </div>
    <div class="center-tree">
      <div class="tree-filter-bar">
        <input type="text" id="tree-search" class="search-input" placeholder="Filter files by path (e.g. /etc, .js, .env)..." oninput="renderFileTree()">
        <label style="display:flex; align-items:center; gap:0.5rem; font-size:0.8125rem; cursor:pointer;">
          <input type="checkbox" id="wasted-toggle" onchange="renderFileTree()">
          <span style="color:var(--purple); font-weight:600;">Wasted files only</span>
        </label>
      </div>
      <div class="tree-scroll" id="tree-container">
        <!-- Rendered by JS -->
      </div>
    </div>
  </div>

  <!-- 2. Diff View -->
  <div id="view-diff" class="view-panel" style="flex-direction:column; overflow-y:auto; padding:1.5rem;">
    <div style="background:var(--bg-card); border:1px solid var(--border); border-radius:8px; padding:1.25rem; margin-bottom:1.5rem;">
      <h3 style="margin-bottom:0.75rem;">Compare Image Tags or Architectures</h3>
      <div style="display:flex; gap:1rem; align-items:flex-end;">
        <div style="flex:1;">
          <label style="font-size:0.8125rem; color:var(--text-muted);">Base Image (A)</label>
          <input type="text" id="diff-img-a" class="search-input" style="width:100%%; margin-top:0.25rem;" value="%s">
        </div>
        <div style="flex:1;">
          <label style="font-size:0.8125rem; color:var(--text-muted);">Target Image (B)</label>
          <input type="text" id="diff-img-b" class="search-input" style="width:100%%; margin-top:0.25rem;" placeholder="e.g. myapp:v1.1.0 or ghcr.io/org/repo:latest">
        </div>
        <button class="btn" onclick="runDiff()">Compare Side-by-Side</button>
      </div>
    </div>
    <div id="diff-results-container">
      <p style="color:var(--text-muted);">Select two images or architecture variants above and click Compare to calculate byte-level file deltas.</p>
    </div>
  </div>

  <!-- 3. Security View -->
  <div id="view-security" class="view-panel" style="flex-direction:column; overflow-y:auto;">
    <div class="card-grid">
      <div class="stat-card">
        <div class="stat-title">Total Leaks Detected</div>
        <div class="stat-value" style="color:var(--red);" id="sec-total">%d</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">Intermediate Layer Leaks</div>
        <div class="stat-value" style="color:#f87171;" id="sec-intermediate">%d</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">Container Root User</div>
        <div class="stat-value" id="sec-root">Auditing...</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">SUID/SGID Binaries</div>
        <div class="stat-value" id="sec-suid">0</div>
      </div>
    </div>
    <div class="content-table-wrapper">
      <h3 style="margin-bottom:1rem;">Detected Credentials & Tokens</h3>
      <table class="data-table">
        <thead>
          <tr><th>Severity</th><th>Rule</th><th>File Path</th><th>Layer</th><th>Intermediate Leak</th><th>Masked Secret</th></tr>
        </thead>
        <tbody id="sec-table-body">
          <!-- Rendered by JS -->
        </tbody>
      </table>
    </div>
  </div>

  <!-- 4. SBOM View -->
  <div id="view-sbom" class="view-panel" style="flex-direction:column; overflow-y:auto;">
    <div style="padding:1.5rem 1.5rem 0.5rem 1.5rem; display:flex; justify-content:space-between; align-items:center;">
      <div>
        <h3 style="margin-bottom:0.25rem;">Software Bill of Materials (SBOM)</h3>
        <p style="color:var(--text-muted); font-size:0.875rem;">Discovered system packages and language dependencies.</p>
      </div>
      <div style="display:flex; gap:0.5rem;">
        <button class="btn" style="background:#0f172a; border:1px solid var(--border);" onclick="window.open('/api/sbom/cyclonedx')">Export CycloneDX 1.5</button>
        <button class="btn" style="background:#0f172a; border:1px solid var(--border);" onclick="window.open('/api/sbom/spdx')">Export SPDX 2.3</button>
      </div>
    </div>
    <div style="padding:0.75rem 1.5rem;">
      <input type="text" id="sbom-search" class="search-input" style="width:100%%;" placeholder="Search package name, license, version..." oninput="filterSBOM()">
    </div>
    <div class="content-table-wrapper">
      <table class="data-table">
        <thead>
          <tr><th>Type</th><th>Package</th><th>Version</th><th>License</th><th>PURL</th></tr>
        </thead>
        <tbody id="sbom-table-body">
          <!-- Rendered by JS -->
        </tbody>
      </table>
    </div>
  </div>

  <!-- 5. Advisor View -->
  <div id="view-advisor" class="view-panel" style="flex-direction:column; overflow-y:auto; padding:1.5rem;">
    <h3 style="margin-bottom:0.25rem;">Dockerfile Optimization Recommendations</h3>
    <p style="color:var(--text-muted); font-size:0.875rem; margin-bottom:1.5rem;">Automated fixes for container build cache preservation, layer slimming, and security hardening.</p>
    <div id="advisor-cards-container">
      <!-- Rendered by JS -->
    </div>
  </div>
</main>

<script>
  let activeTab = 'inspector';
  let imageData = null;
  let activeLayerIndex = 0;
  let activeLayerTree = null;
  let sbomData = [];

  function switchTab(tabId) {
    activeTab = tabId;
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.view-panel').forEach(panel => panel.classList.remove('active'));
    event.currentTarget.classList.add('active');
    document.getElementById('view-' + tabId).classList.add('active');

    if (tabId === 'security') loadSecurity();
    if (tabId === 'sbom') loadSBOM();
    if (tabId === 'advisor') loadAdvisor();
  }

  async function loadInitialData() {
    try {
      const res = await fetch('/api/image');
      if (!res.ok) return;
      imageData = await res.json();
      renderLayersList();
      selectLayer(0);
    } catch (e) {
      console.error('Failed to load image data', e);
    }
  }

  function renderLayersList() {
    const container = document.getElementById('layer-list-container');
    document.getElementById('layer-count').textContent = imageData.layers.length + ' layers';
    container.innerHTML = '';

    imageData.layers.forEach((layer, idx) => {
      const item = document.createElement('div');
      item.className = 'layer-item' + (idx === activeLayerIndex ? ' selected' : '');
      item.onclick = () => selectLayer(idx);

      const sizeMB = (layer.size / (1024 * 1024)).toFixed(1);
      const wastedTag = layer.wastedBytes > 0 ? '<span class="tag wasted">' + (layer.wastedBytes / 1024).toFixed(0) + 'KB Wasted</span>' : '';

      item.innerHTML = ` + "`" + `
        <div class="layer-item-top">
          <strong style="color:var(--accent);">Layer ${idx}</strong>
          <div>${wastedTag} <span style="font-family:monospace; margin-left:0.25rem;">${sizeMB} MB</span></div>
        </div>
        <div class="layer-cmd">${layer.command || 'No command recorded'}</div>
      ` + "`" + `;
      container.appendChild(item);
    });
  }

  async function selectLayer(idx) {
    activeLayerIndex = idx;
    document.querySelectorAll('.layer-item').forEach((el, i) => {
      el.classList.toggle('selected', i === idx);
    });

    try {
      const res = await fetch('/api/layers/' + idx + '/tree');
      const data = await res.json();
      activeLayerTree = data.nodes;
      renderFileTree();
    } catch (e) {
      console.error(e);
    }
  }

  function renderFileTree() {
    const container = document.getElementById('tree-container');
    if (!activeLayerTree) return;

    const search = document.getElementById('tree-search').value.toLowerCase();
    const wastedOnly = document.getElementById('wasted-toggle').checked;
    container.innerHTML = '';

    const paths = Object.keys(activeLayerTree).sort();
    let renderedCount = 0;

    for (const p of paths) {
      const node = activeLayerTree[p];
      if (search && !p.toLowerCase().includes(search)) continue;
      if (wastedOnly && !node.isWasted && !node.isDir) continue;

      renderedCount++;
      if (renderedCount > 1500) {
        const more = document.createElement('div');
        more.style.padding = '0.5rem';
        more.style.color = 'var(--text-muted)';
        more.textContent = '... (Truncated for performance. Use search filter above)';
        container.appendChild(more);
        break;
      }

      const row = document.createElement('div');
      row.className = 'tree-node-row';

      let changeTag = '';
      if (node.changeType === 'added') changeTag = '<span class="tag added">+ Add</span>';
      else if (node.changeType === 'modified') changeTag = '<span class="tag modified">~ Mod</span>';
      else if (node.changeType === 'deleted') changeTag = '<span class="tag deleted">- Del</span>';

      let wasteTag = '';
      if (node.isWasted) {
        wasteTag = '<span class="tag wasted" title="' + (node.wasteReason || 'Wasted space') + '">* Wasted (' + (node.wastedBytes/1024).toFixed(0) + 'KB)</span>';
      }

      const sizeStr = node.isDir ? '' : (node.size > 1048576 ? (node.size/1048576).toFixed(1) + 'MB' : (node.size/1024).toFixed(0) + 'KB');

      row.innerHTML = ` + "`" + `
        <span class="node-icon">${node.isDir ? '📁' : '📄'}</span>
        <span class="node-path">${node.path}</span>
        <span class="node-size">${sizeStr}</span>
        ${changeTag}
        ${wasteTag}
      ` + "`" + `;
      container.appendChild(row);
    }
  }

  async function loadSecurity() {
    try {
      const res = await fetch('/api/security');
      const data = await res.json();
      document.getElementById('sec-total').textContent = data.totalFindings;
      document.getElementById('sec-intermediate').textContent = data.intermediateLeaks;
      document.getElementById('sec-root').textContent = data.runsAsRoot ? 'YES (UID 0 - Risk)' : 'No (Non-root user)';
      document.getElementById('sec-root').style.color = data.runsAsRoot ? 'var(--red)' : 'var(--green)';
      document.getElementById('sec-suid').textContent = data.suidBinariesCount;

      const tbody = document.getElementById('sec-table-body');
      tbody.innerHTML = '';
      data.findings.forEach(f => {
        const tr = document.createElement('tr');
        tr.innerHTML = ` + "`" + `
          <td><span class="tag deleted">${f.severity}</span></td>
          <td><strong>${f.ruleName}</strong></td>
          <td><code>${f.filePath}</code></td>
          <td>Layer ${f.layerIndex}</td>
          <td>${f.isDeletedInFinal ? '<strong style="color:var(--red);">YES (Deleted in later layer)</strong>' : 'No'}</td>
          <td><code>${f.matchMasked}</code></td>
        ` + "`" + `;
        tbody.appendChild(tr);
      });
    } catch (e) {
      console.error(e);
    }
  }

  async function loadSBOM() {
    try {
      const res = await fetch('/api/sbom');
      const data = await res.json();
      sbomData = data.packages || [];
      renderSBOMTable(sbomData);
    } catch (e) {
      console.error(e);
    }
  }

  function filterSBOM() {
    const q = document.getElementById('sbom-search').value.toLowerCase();
    const filtered = sbomData.filter(p =>
      p.name.toLowerCase().includes(q) ||
      (p.license && p.license.toLowerCase().includes(q)) ||
      p.version.toLowerCase().includes(q)
    );
    renderSBOMTable(filtered);
  }

  function renderSBOMTable(packages) {
    const tbody = document.getElementById('sbom-table-body');
    tbody.innerHTML = '';
    packages.forEach(p => {
      const tr = document.createElement('tr');
      tr.innerHTML = ` + "`" + `
        <td><span class="tag ${p.type.startsWith('os') ? 'added' : 'modified'}">${p.type}</span></td>
        <td><strong>${p.name}</strong></td>
        <td><code>${p.version}</code></td>
        <td>${p.license || 'N/A'}</td>
        <td style="color:var(--text-muted); font-size:0.75rem;"><code>${p.purl || ''}</code></td>
      ` + "`" + `;
      tbody.appendChild(tr);
    });
  }

  async function loadAdvisor() {
    try {
      const res = await fetch('/api/advisor');
      const data = await res.json();
      const container = document.getElementById('advisor-cards-container');
      container.innerHTML = '';

      data.recommendations.forEach(r => {
        const card = document.createElement('div');
        card.className = 'stat-card';
        card.style.marginBottom = '1.25rem';
        const snippet = r.remediationSnippet ? '<pre><code>' + r.remediationSnippet + '</code></pre>' : '';

        card.innerHTML = ` + "`" + `
          <div style="display:flex; justify-content:space-between; align-items:flex-start;">
            <div>
              <h4 style="font-size:1.1rem; margin-bottom:0.25rem;">${r.title}</h4>
              <div style="font-size:0.8125rem; color:var(--text-muted);">${r.category}</div>
            </div>
            <span class="tag ${r.severity === 'high' ? 'deleted' : 'modified'}">${r.severity}</span>
          </div>
          <p style="color:#cbd5e1; margin:0.75rem 0; font-size:0.875rem;">${r.explanation}</p>
          ${snippet}
        ` + "`" + `;
        container.appendChild(card);
      });
    } catch (e) {
      console.error(e);
    }
  }

  async function runDiff() {
    const imgA = document.getElementById('diff-img-a').value;
    const imgB = document.getElementById('diff-img-b').value;
    if (!imgB) {
      alert('Please enter a target image for comparison');
      return;
    }
    const container = document.getElementById('diff-results-container');
    container.innerHTML = '<p style="color:var(--text-muted);">Comparing images and calculating diffs...</p>';

    try {
      const res = await fetch('/api/diff', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ imageA: imgA, imageB: imgB })
      });
      const data = await res.json();
      if (data.error) {
        container.innerHTML = '<p style="color:var(--red);">Error: ' + data.error + '</p>';
        return;
      }

      container.innerHTML = ` + "`" + `
        <div class="card-grid" style="padding:0; margin-bottom:1.5rem;">
          <div class="stat-card">
            <div class="stat-title">Net Size Delta</div>
            <div class="stat-value" style="color:${data.totalSizeDelta > 0 ? 'var(--red)' : 'var(--green)'};">
              ${data.totalSizeDelta > 0 ? '+' : ''}${(data.totalSizeDelta/(1024*1024)).toFixed(2)} MB
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-title">Added Files</div>
            <div class="stat-value" style="color:var(--green);">${data.addedCount}</div>
          </div>
          <div class="stat-card">
            <div class="stat-title">Removed Files</div>
            <div class="stat-value" style="color:var(--red);">${data.removedCount}</div>
          </div>
          <div class="stat-card">
            <div class="stat-title">Modified Files</div>
            <div class="stat-value" style="color:var(--yellow);">${data.modifiedCount}</div>
          </div>
        </div>
        <table class="data-table">
          <thead>
            <tr><th>Status</th><th>File Path</th><th>Delta</th><th>Size in B</th></tr>
          </thead>
          <tbody>
            ${data.files.slice(0, 100).map(f => ` + "`" + `
              <tr>
                <td><span class="tag ${f.status === 'added' ? 'added' : f.status === 'removed' ? 'deleted' : 'modified'}">${f.status}</span></td>
                <td><code>${f.path}</code></td>
                <td>${(f.sizeDelta/1024).toFixed(1)} KB</td>
                <td>${(f.sizeB/1024).toFixed(1)} KB</td>
              </tr>
            ` + "`" + `).join('')}
          </tbody>
        </table>
      ` + "`" + `;
    } catch (e) {
      container.innerHTML = '<p style="color:var(--red);">Failed to calculate diff: ' + e.message + '</p>';
    }
  }

  loadInitialData();
</script>
</body>
</html>`,
		imageName,
		imageName,
		arch,
		totalSizeMB,
		effScore, grade,
		func() string {
			if intermediateCount > 0 {
				return fmt.Sprintf(`<div class="pill danger"><span>Intermediate Leaks:</span> <strong>%d</strong></div>`, intermediateCount)
			}
			return fmt.Sprintf(`<div class="pill warn"><span>Wasted:</span> <strong>%.1f MB</strong></div>`, wastedMB)
		}(),
		secCount,
		imageName,
		secCount,
		intermediateCount,
	)
}
