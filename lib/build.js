#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

function build() {
  try {
    console.log('🔨 Building Asgardeo MCP Server...');
    
    // Check if Go is available
    try {
      execSync('go version', { stdio: 'pipe' });
    } catch (error) {
      throw new Error('Go is not installed or not in PATH');
    }
    
    // Build the Go binary
    const platform = process.platform;
    const extension = platform === 'win32' ? '.exe' : '';
    const binaryName = 'asgardeo-mcp' + extension;
    const binDir = path.join(__dirname, '..', 'bin');
    
    // Create bin directory if it doesn't exist
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    const outputPath = path.join(binDir, binaryName);
    const buildCmd = `go build -o "${outputPath}"`;
    
    console.log(`Running: ${buildCmd}`);
    execSync(buildCmd, { stdio: 'inherit' });
    
    // Make executable on Unix systems
    if (platform !== 'win32') {
      fs.chmodSync(outputPath, 0o755);
    }
    
    console.log('✅ Build completed successfully!');
    console.log(`Binary created at: ${outputPath}`);
    
  } catch (error) {
    console.error('❌ Build failed:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  build();
}

module.exports = { build };