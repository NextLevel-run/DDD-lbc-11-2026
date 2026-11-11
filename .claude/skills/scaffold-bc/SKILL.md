---
name: scaffold-bc
description: This skill should be used when users want to create a new bounded context directory structure following DDD and Hexagonal Architecture patterns. Use when users ask to "create a bounded context", "scaffold a domain", or "setup a new module structure".
---

# Scaffold Bounded Context

Creates the standard directory structure of a bounded context in the Go API.

## How to Use

Run the bundled script from anywhere in the repository:

```bash
bash .claude/skills/scaffold-bc/scripts/scaffold-bc.sh <bounded-context-name>
```

It creates:

```
apps/api/internal/<bounded-context-name>/
├── AGENTS.md              # Five empty sections, see below
├── domain/                # Aggregates, value objects, events, ports
├── application/
│   ├── command/           # Write use cases
│   └── query/             # Read use cases
└── adapter/
    ├── driven/inmemory/   # In-memory implementations of the ports
    └── driving/
        ├── http/          # HTTP handlers
        └── consumer/      # Event consumers
```

## After Scaffolding

1. Add the bounded context to the domain map of `documentation/glossary/domain.md`.
2. Fill the two-line role at the top of the generated `AGENTS.md`. Leave the other sections empty until there is something non-obvious to write: the file holds what the code does not say, never an inventory of it.
3. Implement with the `tactical-ddd` skill.

## Naming Conventions

- Bounded context directory: `kebab-case` (`user-management`, `classified-ad`)
- Files: `snake_case` (`user_account.go`)
- Go packages: `lowercase` (`package domain`)
