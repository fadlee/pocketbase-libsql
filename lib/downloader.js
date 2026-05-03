const https = require('node:https');
const http = require('node:http');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { pipeline } = require('node:stream');
const { promisify } = require('node:util');
const { createHash } = require('node:crypto');

const streamPipeline = promisify(pipeline);
const REPO = 'fadlee/pocketbase-libsql';
const PACKAGE_VERSION = require('../package.json').version;

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;

  let platformName;
  let archName;
  let extension = '';

  if (platform === 'win32') {
    platformName = 'windows';
    extension = '.exe';
  } else if (platform === 'darwin') {
    platformName = 'darwin';
  } else if (platform === 'linux') {
    platformName = 'linux';
  } else {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  if (arch === 'x64') {
    archName = 'amd64';
  } else if (arch === 'arm64') {
    archName = 'arm64';
  } else {
    throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { platformName, archName, extension };
}

function getCacheDir() {
  return path.join(os.homedir(), '.cache', 'pocketbase-libsql-bin');
}

function getRequestedVersion() {
  return String(PACKAGE_VERSION).replace(/^v/, '');
}

function getAssetName(version) {
  const { platformName, archName } = getPlatformInfo();
  const ext = platformName === 'windows' ? '.exe' : '';
  return `pocketbase-libsql-v${version}-${platformName}-${archName}${ext}`;
}

function getDownloadUrl(version) {
  const assetName = getAssetName(version);
  return `https://github.com/${REPO}/releases/download/v${version}/${assetName}`;
}

async function getBinaryInfo() {
  const version = getRequestedVersion();
  const { extension } = getPlatformInfo();
  const cacheDir = getCacheDir();
  const assetName = getAssetName(version);
  const versionDir = path.join(cacheDir, version);
  const binaryName = `pocketbase-libsql${extension}`;
  const binaryPath = path.join(versionDir, binaryName);
  const tempDownloadPath = path.join(versionDir, `${assetName}.download`);

  return {
    version,
    cacheDir,
    versionDir,
    binaryName,
    binaryPath,
    assetName,
    downloadUrl: getDownloadUrl(version),
    tempDownloadPath
  };
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function fileExists(filePath) {
  try {
    fs.accessSync(filePath, fs.constants.F_OK);
    return true;
  } catch {
    return false;
  }
}

function isExecutable(filePath) {
  if (!fileExists(filePath)) {
    return false;
  }

  if (process.platform === 'win32') {
    return true;
  }

  try {
    fs.accessSync(filePath, fs.constants.X_OK);
    return true;
  } catch {
    return false;
  }
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith('https:') ? https : http;

    const request = client.get(url, {
      headers: {
        'User-Agent': 'pocketbase-libsql-bin'
      }
    }, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        return downloadFile(response.headers.location, dest).then(resolve).catch(reject);
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Download failed: ${response.statusCode} ${response.statusMessage}`));
        return;
      }

      const totalSize = Number.parseInt(response.headers['content-length'] || '0', 10);
      let downloadedSize = 0;

      response.on('data', (chunk) => {
        downloadedSize += chunk.length;
        if (totalSize > 0) {
          const progress = ((downloadedSize / totalSize) * 100).toFixed(1);
          process.stdout.write(`\rDownloading pocketbase-libsql... ${progress}%`);
        }
      });

      const fileStream = fs.createWriteStream(dest);
      streamPipeline(response, fileStream)
        .then(() => {
          process.stdout.write('\n');
          resolve();
        })
        .catch(reject);
    });

    request.on('error', reject);
    request.setTimeout(30000, () => {
      request.destroy();
      reject(new Error('Download timeout'));
    });
  });
}

function sha256(filePath) {
  const hash = createHash('sha256');
  const fileBuffer = fs.readFileSync(filePath);
  hash.update(fileBuffer);
  return hash.digest('hex');
}

async function tryVerifyChecksum(versionDir, version, assetName, binaryPath) {
  const checksumUrl = `https://github.com/${REPO}/releases/download/v${version}/checksums.txt`;
  const checksumPath = path.join(versionDir, 'checksums.txt');

  try {
    await downloadFile(checksumUrl, checksumPath);
    const checksums = fs.readFileSync(checksumPath, 'utf8').split(/\r?\n/);
    const line = checksums.find((item) => item.trim().endsWith(` ${assetName}`));
    if (!line) {
      return;
    }

    const expected = line.trim().split(/\s+/)[0];
    const actual = sha256(binaryPath);

    if (expected !== actual) {
      throw new Error(`Checksum mismatch for ${assetName}`);
    }
  } catch (error) {
    if (error.message.startsWith('Checksum mismatch')) {
      throw error;
    }
    console.warn(`Checksum verification skipped: ${error.message}`);
  }
}

async function downloadBinary() {
  const info = await getBinaryInfo();
  ensureDir(info.versionDir);

  console.log(`Using pocketbase-libsql version: ${info.version}`);
  console.log(`Platform: ${process.platform}-${process.arch}`);
  console.log(`URL: ${info.downloadUrl}`);

  await downloadFile(info.downloadUrl, info.tempDownloadPath);
  fs.renameSync(info.tempDownloadPath, info.binaryPath);

  if (process.platform !== 'win32') {
    fs.chmodSync(info.binaryPath, 0o755);
  }

  await tryVerifyChecksum(info.versionDir, info.version, info.assetName, info.binaryPath);

  console.log(`✅ pocketbase-libsql binary ready: ${info.binaryPath}`);
  return info.binaryPath;
}

async function ensureBinary() {
  const info = await getBinaryInfo();
  ensureDir(info.versionDir);

  if (isExecutable(info.binaryPath)) {
    console.log(`✅ Using cached pocketbase-libsql binary (v${info.version})`);
    return info.binaryPath;
  }

  return downloadBinary();
}

module.exports = {
  ensureBinary,
  getBinaryInfo,
  getPlatformInfo
};
