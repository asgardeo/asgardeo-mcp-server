#!/usr/bin/env node

const { install } = require('./install');
const path = require('path');
const fs = require('fs');

async function test() {
  console.log('🧪 Testing Asgardeo MCP Server installation...');
  
  try {
    // Test the install function
    await install();
    
    // Check if binary was created
    const platform = process.platform;
    const extension = platform === 'win32' ? '.exe' : '';
    const binaryPath = path.join(__dirname, '..', 'bin', 'asgardeo-mcp' + extension);
    
    if (fs.existsSync(binaryPath)) {
      console.log('✅ Binary installed successfully');
      
      // Check if binary is executable
      const stats = fs.statSync(binaryPath);
      if (stats.mode & parseInt('111', 8)) {
        console.log('✅ Binary is executable');
      } else {
        console.log('⚠️  Binary may not be executable');
      }
    } else {
      console.log('❌ Binary not found at expected location');
      return false;
    }
    
    console.log('✅ All tests passed!');
    return true;
    
  } catch (error) {
    console.error('❌ Test failed:', error.message);
    return false;
  }
}

if (require.main === module) {
  test().then(success => {
    process.exit(success ? 0 : 1);
  });
}

module.exports = { test };