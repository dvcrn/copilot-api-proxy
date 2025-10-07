#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const { spawnSync } = require('child_process');
const crypto = require('crypto');

const OWNER = 'dvcrn';
const REPO = 'copilot-api-proxy';

function mapOs(osName) {
  if (osName === 'win32') return 'windows';
  if (osName === 'darwin') return 'darwin';
  if (osName === 'linux') return 'linux';
  return osName;
}

function mapArch(arch) {
  if (arch === 'x64') return 'amd64';
  if (arch === 'arm64') return 'arm64';
  return arch;
}

function httpGet(url, { headers } = {}) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, { headers }, (res) => {
      if (res.statusCode && res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        // follow redirects
        resolve(httpGet(res.headers.location, { headers }));
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`GET ${url} -> ${res.statusCode}`));
        return;
      }
      const chunks = [];
      res.on('data', (c) => chunks.push(c));
      res.on('end', () => resolve(Buffer.concat(chunks)));
    });
    req.on('error', reject);
  });
}

function sha256(buf) {
  const h = crypto.createHash('sha256');
  h.update(buf);
  return h.digest('hex');
}

(async function main() {
  try {
    const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, 'package.json'), 'utf8'));
    const version = (process.env.COPILOT_PROXY_VERSION || pkg.version || '').replace(/^v/, '');
    if (!version) {
      console.error('postinstall: could not determine version from package.json');
      process.exit(1);
    }
    const osName = mapOs(process.platform);
    const archName = mapArch(process.arch);
    if (!['windows', 'darwin', 'linux'].includes(osName) || !['amd64', 'arm64'].includes(archName)) {
      console.error(`Unsupported platform: os=${osName}, arch=${archName}`);
      process.exit(1);
    }

    const assetName = `copilot-proxy_${version}_${osName}_${archName}.tar.gz`;
    const base = process.env.COPILOT_PROXY_BASE_URL || `https://github.com/${OWNER}/${REPO}/releases/download/v${version}`;
    const url = `${base}/${assetName}`;
    const checksumsUrl = `${base}/checksums.txt`;

    const headers = {};
    const token = process.env.GITHUB_TOKEN;
    if (token) headers['Authorization'] = `Bearer ${token}`;
    headers['User-Agent'] = `${REPO}-postinstall`;

    const outDir = __dirname; // package dir

    // If binary already present, skip
    const exe = process.platform === 'win32' ? 'copilot-api-proxy.exe' : 'copilot-api-proxy';
    const binPath = path.join(outDir, exe);
    if (fs.existsSync(binPath)) {
      // already installed
      try { if (process.platform !== 'win32') fs.chmodSync(binPath, 0o755); } catch {}
      console.log(`postinstall: binary already present at ${binPath}, skipping download.`);
      return;
    }

    console.log(`postinstall: downloading ${assetName} from ${url}`);
    const tarGz = await httpGet(url, { headers });

    // Optional checksum verification unless opted out
    if (process.env.COPILOT_PROXY_SKIP_CHECKSUM !== '1') {
      try {
        const checksumsBuf = await httpGet(checksumsUrl, { headers });
        const sumExpected = checksumsBuf
          .toString('utf8')
          .split(/\r?\n/)
          .map((l) => l.trim())
          .filter(Boolean)
          .map((l) => l.split(/[\s\t]+/))
          .find(([, name]) => name === assetName)?.[0];
        if (!sumExpected) throw new Error('asset not found in checksums.txt');
        const sumActual = sha256(tarGz);
        if (sumActual.toLowerCase() !== sumExpected.toLowerCase()) {
          throw new Error(`checksum mismatch: expected ${sumExpected}, got ${sumActual}`);
        }
        console.log('postinstall: checksum OK');
      } catch (e) {
        console.warn(`postinstall: checksum verification skipped/failed: ${e.message}`);
      }
    }

    // Write tarball to temp and extract only the binary to avoid colliding with package files
    const tmpFile = path.join(os.tmpdir(), `${REPO}-${Date.now()}.tar.gz`);
    fs.writeFileSync(tmpFile, tarGz);

    const tarArgs = ['-xzf', tmpFile, '-C', outDir, exe];
    const tarRes = spawnSync('tar', tarArgs, { stdio: 'inherit' });
    if (tarRes.status !== 0) {
      console.error('postinstall: failed to extract binary from tarball; ensure "tar" is available and archive layout matches.');
      try { fs.unlinkSync(tmpFile); } catch {}
      process.exit(1);
    }

    // Ensure executable
    try { if (process.platform !== 'win32') fs.chmodSync(binPath, 0o755); } catch {}
    // Cleanup
    try { fs.unlinkSync(tmpFile); } catch {}

    console.log(`postinstall: installed ${exe} to ${outDir}`);
  } catch (err) {
    console.error(`postinstall error: ${err.message}`);
    process.exit(1);
  }
})();
