const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const http = require('http');
const puppeteer = require('puppeteer-core');

const PORT = 50077;
const BASE_URL = `http://localhost:${PORT}`;
const OUTPUT_DIR = path.join(__dirname, '..', 'docs', 'images');
const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function checkHealth() {
  return new Promise((resolve) => {
    http
      .get(`${BASE_URL}/api/health`, (res) => {
        resolve(res.statusCode === 200);
      })
      .on('error', () => {
        resolve(false);
      });
  });
}

async function waitForServer(timeoutMs = 15000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (await checkHealth()) return true;
    await sleep(400);
  }
  return false;
}

async function main() {
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  console.log(`Starting LayerScope demo studio server on port ${PORT}...`);
  const binPath = path.join(__dirname, '..', 'bin', 'layerscope.exe');
  const serverProc = spawn(binPath, ['analyze', '--demo', '--no-browser', '-p', String(PORT)], {
    stdio: 'ignore',
    detached: false,
  });

  const ready = await waitForServer();
  if (!ready) {
    serverProc.kill();
    throw new Error('Server did not start within timeout');
  }
  console.log('✓ LayerScope Studio server is healthy and ready.');

  console.log('Launching headless Chrome via puppeteer-core...');
  const browser = await puppeteer.launch({
    executablePath: CHROME_PATH,
    headless: 'new',
    defaultViewport: {
      width: 1600,
      height: 960,
      deviceScaleFactor: 2, // HiDPI Retina quality
    },
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
  });

  const page = await browser.newPage();
  await page.goto(BASE_URL, { waitUntil: 'networkidle0' });
  await sleep(1500); // Allow Monaco/CSS animations to settle

  // 1. Screenshot: Main Layer Inspector
  console.log('Capturing Layer Inspector...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-inspector.png'),
  });

  // 2. Click on a file in the file tree to open the Inspection Drawer
  console.log('Opening File Preview Drawer...');
  const clickedFile = await page.evaluate(() => {
    // Find clickable row containing /etc/os-release or any file path
    const fileSpans = Array.from(document.querySelectorAll('div')).filter(
      (el) => el.style && el.style.cursor === 'pointer' && el.textContent && el.textContent.includes('/etc/os-release')
    );
    if (fileSpans.length > 0) {
      fileSpans[0].click();
      return true;
    }
    // Fallback: any clickable file row in tree
    const allFileRows = Array.from(document.querySelectorAll('div')).filter(
      (el) => el.style && el.style.cursor === 'pointer'
    );
    if (allFileRows.length > 0) {
      allFileRows[0].click();
      return true;
    }
    return false;
  });
  console.log(`File drawer click result: ${clickedFile}`);

  await sleep(1500); // Allow drawer transition and syntax highlighting
  console.log('Capturing File Inspection Drawer...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-preview-drawer.png'),
  });

  // Close the drawer if it has a close button
  await page.evaluate(() => {
    const closeBtns = Array.from(document.querySelectorAll('button')).filter(
      (b) => b.querySelector('svg') && b.closest('[style*="position: fixed"]')
    );
    if (closeBtns.length > 0) closeBtns[0].click();
  });
  await sleep(500);

  // 3. Open the "Switch Image" modal
  console.log('Opening Switch Image Modal...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button')).filter(
      (b) => b.textContent && b.textContent.includes('Switch Image')
    );
    if (buttons.length > 0) buttons[0].click();
  });
  await sleep(800);
  console.log('Capturing Switch Image Modal...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-switcher.png'),
  });

  // Close modal by clicking Cancel or overlay
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button')).filter(
      (b) => b.textContent && b.textContent.trim() === 'Cancel'
    );
    if (buttons.length > 0) buttons[0].click();
  });
  await sleep(600);

  // Helper function to switch tabs
  const switchTab = async (tabName) => {
    return await page.evaluate((name) => {
      const navButtons = Array.from(document.querySelectorAll('nav button'));
      for (const b of navButtons) {
        if (b.textContent && b.textContent.toLowerCase().includes(name.toLowerCase())) {
          b.click();
          return b.textContent.trim();
        }
      }
      return null;
    }, tabName);
  };

  // 4. Tab: Secret Audit
  console.log('Navigating to Secret Audit...');
  const secTab = await switchTab('Secret Audit');
  console.log(`Switched to: ${secTab}`);
  await sleep(1200);
  console.log('Capturing Secret Auditor...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-secrets.png'),
  });

  // 5. Tab: SBOM Packages
  console.log('Navigating to SBOM Packages...');
  const sbomTab = await switchTab('SBOM Packages');
  console.log(`Switched to: ${sbomTab}`);
  await sleep(1200);
  console.log('Capturing SBOM & CVE Tracker...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-sbom.png'),
  });

  // 6. Tab: Optimization Advisor
  console.log('Navigating to Optimization Advisor...');
  const advTab = await switchTab('Optimization Advisor');
  console.log(`Switched to: ${advTab}`);
  await sleep(1500);
  console.log('Capturing Optimization Advisor...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-advisor.png'),
  });

  // 7. Tab: Side-by-Side Diff
  console.log('Navigating to Side-by-Side Diff...');
  const diffTab = await switchTab('Side-by-Side Diff');
  console.log(`Switched to: ${diffTab}`);
  await sleep(1000);

  // Type comparison target 'demo' into React input
  const compInput = await page.$('input[placeholder*="myapp"]');
  if (compInput) {
    await compInput.type('demo', { delay: 50 });
    await sleep(500);
  }

  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const compareBtn = buttons.find((b) => b.textContent && b.textContent.includes('Compare'));
    if (compareBtn) compareBtn.click();
  });
  await sleep(2500); // Allow diff calculation and table rendering

  console.log('Capturing Tag & Arch Comparator...');
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'layerscope-diff.png'),
  });

  await browser.close();
  try {
    if (process.platform === 'win32' && serverProc.pid) {
      execSync(`taskkill /pid ${serverProc.pid} /f /t`, { stdio: 'ignore' });
    } else {
      serverProc.kill();
    }
  } catch (_) {}
  console.log('✓ All authentic screenshots captured successfully in docs/images/');

  // 8. Generate Animated Demo GIF using ffmpeg
  console.log('Synthesizing animated demo GIF with ffmpeg...');
  const frames = [
    'layerscope-inspector.png',
    'layerscope-preview-drawer.png',
    'layerscope-secrets.png',
    'layerscope-sbom.png',
    'layerscope-advisor.png',
    'layerscope-diff.png',
  ];

  const concatListPath = path.join(OUTPUT_DIR, 'gif_frames.txt');
  let concatContent = '';
  for (const f of frames) {
    const fullPath = path.join(OUTPUT_DIR, f).replace(/\\/g, '/');
    concatContent += `file '${fullPath}'\nduration 2.2\n`;
  }
  concatContent += `file '${path.join(OUTPUT_DIR, frames[frames.length - 1]).replace(/\\/g, '/')}'\n`;
  fs.writeFileSync(concatListPath, concatContent);

  const gifPath = path.join(OUTPUT_DIR, 'layerscope_demo.gif');
  const ffmpegCmd = `ffmpeg -y -f concat -safe 0 -i "${concatListPath}" -vf "fps=10,scale=1200:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128:stats_mode=diff[p];[s1][p]paletteuse=dither=bayer:bayer_scale=3" "${gifPath}"`;

  try {
    execSync(ffmpegCmd, { stdio: 'ignore' });
    console.log(`✓ Animated demo GIF created at ${gifPath}`);
  } catch (err) {
    console.error('Error generating GIF with ffmpeg:', err.message);
  } finally {
    if (fs.existsSync(concatListPath)) {
      fs.unlinkSync(concatListPath);
    }
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
