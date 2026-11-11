#!/usr/bin/env node

// Size budget of the agent docs. An AGENTS.md is a map plus the rules an agent
// would break, not an inventory of the code: past a few thousand tokens agents
// stop reading it. A glossary entry says what a term is in one or two
// sentences; the rules of a feature belong to its module's AGENTS.md.
//
// Usage:
//   node tools/docs-budget/docs-budget.mjs [--root <path>] [file ...]
//
// Without files, every AGENTS.md and both glossaries are checked (`make check`
// and CI). With files (the agent hook passes the edited ones), only the
// AGENTS.md and glossaries among them are checked.
//
// Exit codes: 0 within budget, 1 a violation, 3 the check itself failed.

import { readFile, readdir } from 'node:fs/promises'
import { basename, isAbsolute, join, relative, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

export const BUDGET_BYTES = 16 * 1024
const DOC_NAME = 'AGENTS.md'
const SKIPPED_DIRECTORIES = new Set(['.git', 'node_modules', 'dist', '.output', '.tanstack', 'worktrees'])

export const GLOSSARIES = new Map([
  ['documentation/glossary/domain.md', { bytes: 20 * 1024, entryCharacters: 300 }],
  ['documentation/glossary/technical.md', { bytes: 20 * 1024, entryCharacters: 300 }],
])

// The verdict on one AGENTS.md, `null` when it is within its budget.
export function judge({ path, size, budget = BUDGET_BYTES }) {
  return size > budget ? `${path}: ${size} bytes, over the ${budget}-byte budget` : null
}

// The entries of a glossary, `- **Term** …` at any depth: the bullet line and
// the lines indented under it, joined onto one line. A nested bullet starts an
// entry of its own and ends its parent's text.
export function glossaryEntries(text) {
  const entries = []
  let current = null
  for (const [index, line] of text.split('\n').entries()) {
    const bullet = /^(\s*)- (.*)$/.exec(line)
    if (bullet) {
      const term = /^\*\*(.+?)\*\*/.exec(bullet[2])
      current = term ? { term: term[1], line: index + 1, text: bullet[2] } : null
      if (current) entries.push(current)
    } else if (current && line.trim() !== '' && /^\s/.test(line)) {
      current.text += ` ${line.trim()}`
    } else {
      current = null
    }
  }
  return entries.map(({ term, line, text }) => ({ term, line, length: [...text].length }))
}

// The verdicts on one glossary, empty when it is within its budget.
export function judgeGlossary({ path, text, rule }) {
  const violations = []
  const size = Buffer.byteLength(text)
  if (size > rule.bytes) {
    violations.push(`${path}: ${size} bytes, over the ${rule.bytes}-byte budget`)
  }
  for (const entry of glossaryEntries(text)) {
    if (entry.length > rule.entryCharacters) {
      violations.push(
        `${path}:${entry.line}: entry "${entry.term}" is ${entry.length} characters, over the ${rule.entryCharacters}-character limit`,
      )
    }
  }
  return violations
}

async function everyDoc(root, directory = '') {
  const docs = []
  for (const entry of await readdir(join(root, directory), { withFileTypes: true })) {
    const path = directory === '' ? entry.name : `${directory}/${entry.name}`
    if (entry.isDirectory()) {
      if (!SKIPPED_DIRECTORIES.has(entry.name)) docs.push(...(await everyDoc(root, path)))
    } else if (entry.name === DOC_NAME) {
      docs.push(path)
    }
  }
  return docs
}

async function textOf(root, path) {
  try {
    return await readFile(resolve(root, path), 'utf8')
  } catch (error) {
    if (error.code === 'ENOENT') return null
    throw error
  }
}

function toRepositoryPath(root, file) {
  const path = isAbsolute(file) ? relative(root, file) : file
  return path.split(sep).join('/')
}

export async function run({ root, files }) {
  const paths = files.map((file) => toRepositoryPath(root, file))
  const full = paths.length === 0
  const docs = full ? await everyDoc(root) : paths.filter((path) => basename(path) === DOC_NAME)
  const glossaries = full ? [...GLOSSARIES.keys()] : paths.filter((path) => GLOSSARIES.has(path))

  const violations = []
  for (const path of docs) {
    const text = await textOf(root, path)
    if (text === null) continue
    const verdict = judge({ path, size: Buffer.byteLength(text) })
    if (verdict) violations.push(verdict)
  }
  for (const path of glossaries) {
    const text = await textOf(root, path)
    if (text === null) continue
    violations.push(...judgeGlossary({ path, text, rule: GLOSSARIES.get(path) }))
  }
  return violations
}

async function main() {
  const args = process.argv.slice(2)
  const rootIndex = args.indexOf('--root')
  if (rootIndex !== -1 && !args[rootIndex + 1]) throw new Error('--root requires a path')
  const root = resolve(
    rootIndex === -1 ? resolve(fileURLToPath(import.meta.url), '..', '..', '..') : args[rootIndex + 1],
  )
  const files = args.filter(
    (_, index) => rootIndex === -1 || (index !== rootIndex && index !== rootIndex + 1),
  )

  const violations = await run({ root, files })
  if (violations.length === 0) return 0
  for (const violation of violations) console.error(`docs-budget: ${violation}`)
  const isGlossary = (violation) =>
    [...GLOSSARIES.keys()].some((path) => violation.startsWith(`${path}:`))
  if (violations.some((violation) => !isGlossary(violation))) {
    console.error(
      'docs-budget: an AGENTS.md is a map plus the rules an agent would break; drop what ls or grep finds, history and paraphrase of the code',
    )
  }
  if (violations.some(isGlossary)) {
    console.error(
      "docs-budget: a glossary entry says what a term is, in one or two sentences; a feature's rules belong to its module's AGENTS.md",
    )
  }
  return 1
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().then(
    (code) => process.exit(code),
    (error) => {
      console.error(`docs-budget: ${error.message}`)
      process.exit(3)
    },
  )
}
