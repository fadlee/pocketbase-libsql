#!/usr/bin/env node

const { spawn } = require('node:child_process');
const { ensureBinary } = require('../lib/downloader');

function printWrapperHelp() {
  console.log(`
PocketBase libSQL npm wrapper

This package version matches repository release version and downloads matching binary automatically.

Examples:
  npx pocketbase-libsql-bin serve
  npx pocketbase-libsql-bin --help
`);
}

async function main() {
  try {
    const args = process.argv.slice(2);

    if (args.includes('--help') || args.includes('-h')) {
      printWrapperHelp();
    }

    const binaryPath = await ensureBinary();

    const child = spawn(binaryPath, args, {
      stdio: 'inherit',
      shell: false,
      cwd: process.cwd(),
      env: process.env
    });

    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
        return;
      }
      process.exit(code ?? 0);
    });

    child.on('error', (err) => {
      console.error('Failed to start pocketbase-libsql binary:', err);
      process.exit(1);
    });

    process.on('SIGINT', () => child.kill('SIGINT'));
    process.on('SIGTERM', () => child.kill('SIGTERM'));
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

main();
