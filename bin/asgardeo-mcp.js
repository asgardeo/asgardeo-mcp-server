#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

function getBinaryPath() {
  const platform = process.platform;
  const extension = platform === 'win32' ? '.exe' : '';
  const binaryName = 'asgardeo-mcp' + extension;
  
  // Try to find the binary in the bin directory
  const binPath = path.join(__dirname, binaryName);
  
  if (fs.existsSync(binPath)) {
    return binPath;
  }
  
  // Fallback: try to use system PATH
  return 'asgardeo-mcp';
}

function main() {
  const binaryPath = getBinaryPath();
  const args = process.argv.slice(2);
  
  // Spawn the Go binary with all arguments and environment variables
  const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    env: process.env
  });
  
  child.on('error', (err) => {
    if (err.code === 'ENOENT') {
      console.error('❌ Asgardeo MCP Server binary not found.');
      console.error('Please run: npm install @asgardeo/mcp-server');
      process.exit(1);
    } else {
      console.error('❌ Failed to start Asgardeo MCP Server:', err.message);
      process.exit(1);
    }
  });
  
  child.on('exit', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exit(code || 0);
    }
  });
}

main();