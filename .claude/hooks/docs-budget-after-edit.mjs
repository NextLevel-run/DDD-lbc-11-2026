#!/usr/bin/env node

// PostToolUse hook: right after an agent edits an AGENTS.md or a glossary,
// tells it when the file went over its budget, so it compacts the file in the
// same turn. Exit 2 hands the verdict to the agent; an internal failure never
// blocks editing.

import { execFileSync } from 'node:child_process'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')

let payload = ''
for await (const chunk of process.stdin) payload += chunk

let file
try {
  file = JSON.parse(payload).tool_input?.file_path
} catch {
  process.exit(0)
}
if (!file || !/(^|[\\/])AGENTS\.md$|glossary[\\/][^\\/]+\.md$/.test(file)) process.exit(0)

try {
  execFileSync('node', [resolve(root, 'tools/docs-budget/docs-budget.mjs'), '--root', root, file], {
    stdio: ['ignore', 'ignore', 'pipe'],
  })
} catch (error) {
  if (error.status === 1) {
    process.stderr.write(error.stderr)
    process.exit(2)
  }
  console.error(`docs-budget-after-edit: internal error (exit ${error.status})`)
}
process.exit(0)
