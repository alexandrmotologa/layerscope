const fs = require('fs');
const path = require('path');
const { Resvg } = require('@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="cyan-accent" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#38bdf8"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="slate-plate" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>

    <linearGradient id="dark-plate" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#1e293b"/>
      <stop offset="100%" stop-color="#090d16"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#000000" flood-opacity="0.18" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="40" stroke-linejoin="round" />

      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.45" stroke-dasharray="16, 12" />

      <!-- Corner Alignment Crosshairs -->
      <line x1="-338" y1="-195" x2="-300" y2="-195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="338" y1="-195" x2="300" y2="-195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="-338" y1="195" x2="-300" y2="195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="338" y1="195" x2="300" y2="195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>

      <!-- 100% Watertight Solid Base Silhouette -->
      <path d="
        M 0,-290
        L 85,-290 L 155,-325 L 140,-240 L 210,-170 L 225,-70 L 205,50 L 180,180 L 130,280 L 0,335
        L -130,280 L -180,180 L -205,50 L -225,-70 L -210,-170 L -140,-240 L -155,-325 L -85,-290 Z
      " fill="#0b0f19" />

      <!-- Crown & Ear Tufts (Symmetrical low-poly facets) -->
      <!-- Left Ear Tuft -->
      <polygon points="0,-290 -85,-290 -140,-240 0,-240" fill="#1e293b" />
      <polygon points="-85,-290 -155,-325 -140,-240" fill="#334155" />
      <polygon points="-85,-290 -155,-325 -115,-275" fill="#475569" />
      <!-- Right Ear Tuft -->
      <polygon points="0,-290 85,-290 140,-240 0,-240" fill="#0f172a" />
      <polygon points="85,-290 155,-325 140,-240" fill="#1e293b" />
      <polygon points="85,-290 155,-325 115,-275" fill="#334155" />

      <!-- Forehead & Upper Facial Discs -->
      <polygon points="0,-290 -70,-240 0,-210" fill="#475569" />
      <polygon points="0,-290 70,-240 0,-210" fill="#334155" />
      <polygon points="-70,-240 -140,-240 -150,-170 -60,-175" fill="#1e293b" />
      <polygon points="70,-240 140,-240 150,-170 60,-175" fill="#0f172a" />

      <!-- Central Brow & Nose Bridge -->
      <polygon points="0,-210 -60,-175 0,-130" fill="#64748b" />
      <polygon points="0,-210 60,-175 0,-130" fill="#475569" />
      <polygon points="0,-130 -35,-85 0,-40" fill="#334155" />
      <polygon points="0,-130 35,-85 0,-40" fill="#1e293b" />

      <!-- Left Outer Temple / Cheek -->
      <polygon points="-140,-240 -210,-170 -150,-170" fill="#0f172a" />
      <polygon points="-210,-170 -225,-70 -160,-80 -150,-170" fill="#1e293b" />
      <polygon points="-150,-170 -160,-80 -60,-105 -60,-175" fill="#334155" />
      <!-- Right Outer Temple / Cheek -->
      <polygon points="140,-240 210,-170 150,-170" fill="#090d16" />
      <polygon points="210,-170 225,-70 160,-80 150,-170" fill="#0f172a" />
      <polygon points="150,-170 160,-80 60,-105 60,-175" fill="#1e293b" />

      <!-- Ocular Sockets (Angular Night-Vision Gaze) -->
      <!-- Left Eye Socket -->
      <polygon points="-60,-175 -150,-170 -160,-80 -60,-105" fill="#090d16" stroke="#00f5ff" stroke-width="2" stroke-opacity="0.3" />
      <!-- Left Eye Iris (Scope / Lens Aperture) -->
      <polygon points="-75,-140 -125,-145 -140,-115 -105,-100 -70,-115" fill="#0b0f19" stroke="#00f5ff" stroke-width="4" />
      <polygon points="-85,-132 -118,-135 -128,-115 -105,-107 -82,-116" fill="url(#cyan-glow)" />
      <circle cx="-105" cy="-120" r="10" fill="#ffffff" />
      <circle cx="-100" cy="-124" r="4" fill="#ffffff" />

      <!-- Right Eye Socket -->
      <polygon points="60,-175 150,-170 160,-80 60,-105" fill="#060910" stroke="#00f5ff" stroke-width="2" stroke-opacity="0.3" />
      <!-- Right Eye Iris (Scope / Lens Aperture) -->
      <polygon points="75,-140 125,-145 140,-115 105,-100 70,-115" fill="#0b0f19" stroke="#00f5ff" stroke-width="4" />
      <polygon points="85,-132 118,-135 128,-115 105,-107 82,-116" fill="url(#cyan-glow)" />
      <circle cx="105" cy="-120" r="10" fill="#ffffff" />
      <circle cx="100" cy="-124" r="4" fill="#ffffff" />

      <!-- Sharp Obsidian Beak -->
      <polygon points="0,-40 -28,-30 0,35" fill="#1e293b" />
      <polygon points="0,-40 28,-30 0,35" fill="#0b0f19" />
      <polygon points="0,-40 -12,-15 0,35" fill="#334155" />
      <polygon points="0,-40 12,-15 0,35" fill="#1e293b" />
      <!-- Beak Highlight Tip -->
      <polygon points="0,15 -6,26 0,35 6,26" fill="#00f5ff" />

      <!-- Lower Facial Disc & Throat -->
      <polygon points="-60,-105 -160,-80 -120,-10 -35,-85" fill="#1e293b" />
      <polygon points="60,-105 160,-80 120,-10 35,-85" fill="#0f172a" />
      <polygon points="-35,-85 -120,-10 -65,30 0,0" fill="#334155" />
      <polygon points="35,-85 120,-10 65,30 0,0" fill="#1e293b" />
      <polygon points="0,0 -65,30 0,60" fill="#475569" />
      <polygon points="0,0 65,30 0,60" fill="#334155" />

      <!-- OCI Container Stratified Layer Plates (Chest Armor) -->
      <!-- Layer 1 (Top Shield Layer) -->
      <polygon points="0,65 -95,45 -135,90 0,115" fill="#334155" />
      <polygon points="0,65 95,45 135,90 0,115" fill="#1e293b" />
      <polyline points="-135,90 0,115 135,90" fill="none" stroke="#00f5ff" stroke-width="3" stroke-opacity="0.7" />

      <!-- Layer 2 (Middle Shield Layer) -->
      <polygon points="0,120 -115,98 -150,150 0,175" fill="#1e293b" />
      <polygon points="0,120 115,98 150,150 0,175" fill="#0f172a" />
      <polyline points="-150,150 0,175 150,150" fill="none" stroke="#00f5ff" stroke-width="3" stroke-opacity="0.8" />

      <!-- Layer 3 (Lower Base Layer) -->
      <polygon points="0,180 -125,155 -160,215 0,240" fill="#334155" />
      <polygon points="0,180 125,155 160,215 0,240" fill="#1e293b" />
      <polyline points="-160,215 0,240 160,215" fill="none" stroke="#00f5ff" stroke-width="3.5" />

      <!-- Layer 4 / Keel (Bottom Anchor Layer) -->
      <polygon points="0,245 -100,222 -115,285 0,315" fill="#1e293b" />
      <polygon points="0,245 100,222 115,285 0,315" fill="#0b0f19" />
      <polyline points="-115,285 0,315 115,285" fill="none" stroke="#38bdf8" stroke-width="4" />

      <!-- Outer Flank Wings / Armor Guards -->
      <!-- Left Wing Flank -->
      <polygon points="-225,-70 -205,50 -160,-10" fill="#0f172a" />
      <polygon points="-205,50 -180,180 -150,110" fill="#1e293b" />
      <polygon points="-180,180 -130,280 -120,200" fill="#0f172a" />
      <!-- Right Wing Flank -->
      <polygon points="225,-70 205,50 160,-10" fill="#090d16" />
      <polygon points="205,50 180,180 150,110" fill="#0f172a" />
      <polygon points="180,180 130,280 120,200" fill="#090d16" />

      <!-- Core Scope Pulse Ring on Sternum (Negative Space Sensor) -->
      <circle cx="0" cy="180" r="16" fill="#0b0f19" stroke="#00f5ff" stroke-width="4" />
      <circle cx="0" cy="180" r="6" fill="#00f5ff" />
      <line x1="-30" y1="180" x2="-18" y2="180" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="18" y1="180" x2="30" y2="180" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="0" y1="150" x2="0" y2="162" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="0" y1="198" x2="0" y2="210" stroke="#00f5ff" stroke-width="2.5" />

    </g>
  </g>
</svg>`;
}

async function main() {
  const outputDir = path.join(__dirname, '..', 'docs', 'images');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: 1024 },
  });
  const pngData = resvg.render().asPng();
  fs.writeFileSync(pngPath, pngData);

  console.log(`✓ Successfully rendered:`);
  console.log(`  - ${svgPath}`);
  console.log(`  - ${pngPath} (1024x1024)`);
}

main().catch(console.error);
