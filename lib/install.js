#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const GITHUB_REPO = 'asgardeo/asgardeo-mcp-server';
const BINARY_NAME = 'asgardeo-mcp';

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;
  
  let osName, archName, extension = '';
  
  switch (platform) {
    case 'darwin':
      osName = 'darwin';
      break;
    case 'linux':
      osName = 'linux';
      break;
    case 'win32':
      osName = 'windows';
      extension = '.exe';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }
  
  switch (arch) {
    case 'x64':
      archName = 'amd64';
      break;
    case 'arm64':
      archName = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }
  
  return { osName, archName, extension };
}

function getLatestRelease() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${GITHUB_REPO}/releases/latest`,
      headers: {
        'User-Agent': 'asgardeo-mcp-installer'
      }
    };
    
    https.get(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          const release = JSON.parse(data);
          resolve(release);
        } catch (err) {
          reject(err);
        }
      });
    }).on('error', reject);
  });
}

function downloadBinary(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    
    https.get(url, (res) => {
      if (res.statusCode === 302 || res.statusCode === 301) {
        // Handle redirect
        return downloadBinary(res.headers.location, dest);
      }
      
      res.pipe(file);
      file.on('finish', () => {
        file.close();
        fs.chmodSync(dest, 0o755); // Make executable
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {}); // Delete partial file
      reject(err);
    });
  });
}

async function install() {
  try {
    console.log('Installing Asgardeo MCP Server...');
    
    const { osName, archName, extension } = getPlatformInfo();
    const binDir = path.join(__dirname, '..', 'bin');
    const binaryPath = path.join(binDir, BINARY_NAME + extension);
    
    // Create bin directory if it doesn't exist
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    // Skip download if binary already exists
    if (fs.existsSync(binaryPath)) {
      console.log('Binary already exists, skipping download.');
      return;
    }
    
    console.log(`Detecting platform: ${osName}-${archName}`);
    
    // Get latest release info
    const release = await getLatestRelease();
    const assetName = `${BINARY_NAME}-${osName}-${archName}${extension}`;
    
    // Find the matching asset
    const asset = release.assets.find(a => a.name === assetName);
    if (!asset) {
      throw new Error(`No binary found for platform ${osName}-${archName}`);
    }
    
    console.log(`Downloading ${asset.name}...`);
    await downloadBinary(asset.browser_download_url, binaryPath);
    
    console.log('✅ Asgardeo MCP Server installed successfully!');
    console.log(`Binary location: ${binaryPath}`);
    
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  install();
}

module.exports = { install };