#!/usr/bin/env node

const { spawn } = require('node:child_process');
const { ensureBinary } = require('../lib/downloader');

function stripWrapperArgs(rawArgs) {
  const args = [];

  for (let i = 0; i < rawArgs.length; i += 1) {
    const arg = rawArgs[i];

    if (arg === '--pb-version' || arg === '--pbl-version') {
      i += 1;
      continue;
    }

    args.push(arg);
  }

  return args;
}

function printWrapperHelp() {
  console.log(`
PocketBase libSQL npm wrapper - Additional options:

  --pbl-version <version>  Use specific pocketbase-libsql release version
  --pb-version <version>   Alias for --pbl-version

Environment variables:
  POCKETBASE_LIBSQL_VERSION  Set default wrapper release version to use

Examples:
  npx pocketbase-libsql-bin serve
  npx pocketbase-libsql-bin --pbl-version 0.37.5 serve
  POCKETBASE_LIBSQL_VERSION=0.37.5 npx pocketbase-libsql-bin serve
`);
}

async function main() {
  try {
    const rawArgs = process.argv.slice(2);
    const args = stripWrapperArgs(rawArgs);

    if (rawArgs.includes('--help') || rawArgs.includes('-h')) {
      printWrapperHelp();
    }

    const binaryPath = await ensureBinary(rawArgs);

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
