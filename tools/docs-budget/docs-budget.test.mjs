import assert from 'node:assert/strict'
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import test from 'node:test'

import { BUDGET_BYTES, glossaryEntries, judge, judgeGlossary, run } from './docs-budget.mjs'

async function fixture(files) {
  const root = await mkdtemp(join(tmpdir(), 'docs-budget-'))
  for (const [path, content] of Object.entries(files)) {
    await mkdir(join(root, dirname(path)), { recursive: true })
    await writeFile(join(root, path), typeof content === 'number' ? 'x'.repeat(content) : content)
  }
  return root
}

test('accepts an AGENTS.md at the budget and refuses one byte more', () => {
  assert.equal(judge({ path: 'AGENTS.md', size: BUDGET_BYTES }), null)
  assert.match(judge({ path: 'AGENTS.md', size: BUDGET_BYTES + 1 }), /over the 16384-byte budget/)
})

test('reads a glossary entry with its continuation lines and its nested entries', () => {
  const entries = glossaryEntries(
    ['- **Port** — interface the domain', '  declares.', '  - **Repository** — a port.', '', 'prose'].join('\n'),
  )
  assert.deepEqual(
    entries.map(({ term, line }) => ({ term, line })),
    [
      { term: 'Port', line: 1 },
      { term: 'Repository', line: 3 },
    ],
  )
  assert.equal(entries[0].length, '**Port** — interface the domain declares.'.length)
})

test('refuses a glossary entry over its character limit', () => {
  const text = `- **Long** — ${'x'.repeat(300)}\n- **Short** — fine\n`
  const violations = judgeGlossary({
    path: 'documentation/glossary/domain.md',
    text,
    rule: { bytes: 20 * 1024, entryCharacters: 300 },
  })
  assert.equal(violations.length, 1)
  assert.match(violations[0], /domain\.md:1: entry "Long"/)
})

test('checks every AGENTS.md and both glossaries when no file is given', async () => {
  const root = await fixture({
    'AGENTS.md': 10,
    'apps/api/internal/ads/AGENTS.md': BUDGET_BYTES + 1,
    'node_modules/pkg/AGENTS.md': BUDGET_BYTES + 1,
    'documentation/glossary/domain.md': `- **Long** — ${'x'.repeat(300)}\n`,
  })
  try {
    const violations = await run({ root, files: [] })
    assert.equal(violations.length, 2)
    assert.match(violations[0], /^apps\/api\/internal\/ads\/AGENTS\.md/)
    assert.match(violations[1], /^documentation\/glossary\/domain\.md:1/)
  } finally {
    await rm(root, { recursive: true })
  }
})

test('checks only the docs among the given files', async () => {
  const root = await fixture({
    'AGENTS.md': BUDGET_BYTES + 1,
    'apps/web/AGENTS.md': BUDGET_BYTES + 1,
    'apps/web/src/router.tsx': 10,
  })
  try {
    const violations = await run({
      root,
      files: [join(root, 'apps/web/AGENTS.md'), 'apps/web/src/router.tsx'],
    })
    assert.equal(violations.length, 1)
    assert.match(violations[0], /^apps\/web\/AGENTS\.md/)
  } finally {
    await rm(root, { recursive: true })
  }
})
