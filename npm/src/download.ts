import * as fs from 'fs';
import * as path from 'path';
import * as https from 'https';
import * as crypto from 'crypto';

const GITHUB_REPO = 'pinchtab/pinchtab';

// Read version from package.json at build time
function getVersion(): string {
  const pkgPath = path.join(__dirname, '..', 'package.json');
  const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf-8'));
  return pkg.version;
}

interface Platform {
  os: 'darwin' | 'linux' | 'windows';
  arch: 'x64' | 'arm64';
}

function detectPlatform(): Platform {
  const platform = process.platform as any;
  const arch = process.arch === 'arm64' ? 'arm64' : 'x64';

  const osMap: Record<string, 'darwin' | 'linux' | 'windows'> = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
  };

  const os = osMap[platform];
  if (!os) {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  return { os, arch };
}

function getBinaryName(platform: Platform): string {
  const { os, arch } = platform;
  const archName = arch === 'arm64' ? 'arm64' : 'x64';

  if (os === 'windows') {
    return `pinchtab-${os}-${archName}.exe`;
  }
  return `pinchtab-${os}-${archName}`;
}

function getBinDir(): string {
  return path.join(process.env.HOME || process.env.USERPROFILE || '', '.pinchtab', 'bin');
}

function fetchUrl(url: string): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    https.get(url, (response) => {
      if (response.statusCode === 404) {
        reject(new Error(`Not found: ${url}`));
        return;
      }

      if (response.statusCode !== 200) {
        reject(new Error(`HTTP ${response.statusCode}: ${url}`));
        return;
      }

      const chunks: Buffer[] = [];
      response.on('data', (chunk) => chunks.push(chunk));
      response.on('end', () => resolve(Buffer.concat(chunks)));
      response.on('error', reject);
    }).on('error', reject);
  });
}

async function downloadChecksums(version: string): Promise<Map<string, string>> {
  const url = `https://github.com/${GITHUB_REPO}/releases/download/v${version}/checksums.txt`;

  try {
    const data = await fetchUrl(url);
    const checksums = new Map<string, string>();

    data
      .toString('utf-8')
      .trim()
      .split('\n')
      .forEach((line) => {
        const [hash, filename] = line.split(/\s+/);
        if (hash && filename) {
          checksums.set(filename.trim(), hash.trim());
        }
      });

    return checksums;
  } catch (err) {
    throw new Error(
      `Failed to download checksums: ${(err as Error).message}. ` +
      `Ensure v${version} is released on GitHub with checksums.txt`
    );
  }
}

function verifySHA256(filePath: string, expectedHash: string): boolean {
  const hash = crypto.createHash('sha256');
  const data = fs.readFileSync(filePath);
  hash.update(data);
  const actualHash = hash.digest('hex');
  return actualHash.toLowerCase() === expectedHash.toLowerCase();
}

async function downloadBinary(platform: Platform, version: string): Promise<void> {
  const binaryName = getBinaryName(platform);
  const binDir = getBinDir();
  const binaryPath = path.join(binDir, binaryName);

  // Skip if already exists
  if (fs.existsSync(binaryPath)) {
    console.log(`✓ Pinchtab binary already present: ${binaryPath}`);
    return;
  }

  // Fetch checksums
  console.log(`Downloading Pinchtab ${version} for ${platform.os}-${platform.arch}...`);
  const checksums = await downloadChecksums(version);

  if (!checksums.has(binaryName)) {
    throw new Error(
      `Binary not found in checksums: ${binaryName}. ` +
      `Available: ${Array.from(checksums.keys()).join(', ')}`
    );
  }

  const expectedHash = checksums.get(binaryName)!;
  const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${version}/${binaryName}`;

  // Ensure directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Download binary
  return new Promise((resolve, reject) => {
    console.log(`Downloading from ${downloadUrl}...`);

    const file = fs.createWriteStream(binaryPath);

    https.get(downloadUrl, (response) => {
      if (response.statusCode !== 200) {
        fs.unlink(binaryPath, () => {});
        reject(new Error(`HTTP ${response.statusCode}: ${downloadUrl}`));
        return;
      }

      response.pipe(file);

      file.on('finish', () => {
        file.close();

        // Verify checksum
        if (!verifySHA256(binaryPath, expectedHash)) {
          fs.unlink(binaryPath, () => {});
          reject(
            new Error(
              `Checksum verification failed for ${binaryName}. ` +
              `Download may be corrupted. Please try again.`
            )
          );
          return;
        }

        // Make executable
        fs.chmodSync(binaryPath, 0o755);
        console.log(`✓ Verified and installed: ${binaryPath}`);
        resolve();
      });

      file.on('error', (err) => {
        fs.unlink(binaryPath, () => {});
        reject(err);
      });
    }).on('error', reject);
  });
}

export async function ensureBinary(): Promise<string> {
  const platform = detectPlatform();
  const version = getVersion();

  await downloadBinary(platform, version);

  const binDir = getBinDir();
  const binaryName = getBinaryName(platform);
  return path.join(binDir, binaryName);
}
