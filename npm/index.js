#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

function mapOs(os) {
  if (os === 'win32') return 'windows';
  if (os === 'darwin') return 'darwin';
  if (os === 'linux') return 'linux';
  return os;
}

function mapArch(arch) {
  if (arch === 'x64') return 'amd64';
  if (arch === 'arm64') return 'arm64';
  // We only build amd64 and arm64; others unsupported.
  return arch;
}

function findBinary(baseDir, osName, archName) {
  const exe = process.platform === 'win32' ? 'copilot-api-proxy.exe' : 'copilot-api-proxy';

  // 1) Prefer binary next to this script (archive root case)
  const rootCandidate = path.join(baseDir, exe);
  if (fs.existsSync(rootCandidate)) return rootCandidate;

  // 2) Look for GoReleaser build output directories in dist pattern
  //    e.g. copilot-api-proxy_darwin_arm64_v8.0, copilot-api-proxy_linux_amd64_v1
  let entries = [];
  try {
    entries = fs.readdirSync(baseDir, { withFileTypes: true });
  } catch (_) {
    entries = [];
  }

  const prefix = `copilot-api-proxy_${osName}_${archName}`;
  const dir = entries
    .filter((e) => e.isDirectory() && e.name.startsWith(prefix))
    // If multiple, prefer ones without extra suffix first, else the shortest name
    .sort((a, b) => a.name.length - b.name.length)[0];

  if (dir) {
    const candidate = path.join(baseDir, dir.name, exe);
    if (fs.existsSync(candidate)) return candidate;
  }

  return null;
}

(function main() {
  const osName = mapOs(process.platform);
  const archName = mapArch(process.arch);
  const baseDir = __dirname; // dist root or archive root

  if (!['windows', 'darwin', 'linux'].includes(osName) || !['amd64', 'arm64'].includes(archName)) {
    console.error(`Unsupported platform: os=${osName}, arch=${archName}`);
    process.exit(1);
  }

  const binPath = findBinary(baseDir, osName, archName);
  if (!binPath) {
    console.error(`Could not locate binary for ${osName}/${archName} under ${baseDir}`);
    try {
      const listing = fs.readdirSync(baseDir).join(', ');
      console.error(`Contents of ${baseDir}: ${listing}`);
    } catch (_) {}
    process.exit(1);
  }

  // Ensure executable on unix
  if (process.platform !== 'win32') {
    try { fs.chmodSync(binPath, 0o755); } catch (_) {}
  }

  const child = spawn(binPath, process.argv.slice(2), {
    stdio: 'inherit',
    windowsHide: true,
  });
  child.on('exit', (code, signal) => {
    if (signal) {
      // Map signals to typical exit codes if needed
      process.exit(1);
    } else {
      process.exit(code == null ? 1 : code);
    }
  });
  child.on('error', (err) => {
    console.error(`Failed to start binary: ${err.message}`);
    process.exit(1);
  });
})();
