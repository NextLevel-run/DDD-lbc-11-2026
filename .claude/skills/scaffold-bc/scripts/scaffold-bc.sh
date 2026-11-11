#!/bin/bash

# scaffold-bc.sh - Create a bounded context structure for DDD + Hexagonal Architecture
# Usage: bash scaffold-bc.sh <bounded-context-name>

set -e

if [ -z "$1" ]; then
    echo "Error: Bounded context name is required"
    echo "Usage: bash scaffold-bc.sh <bounded-context-name>"
    echo "Example: bash scaffold-bc.sh user-management"
    exit 1
fi

BC_NAME="$1"
REPO_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
BASE_DIR="${REPO_ROOT}/apps/api/internal/${BC_NAME}"

if [ -d "$BASE_DIR" ]; then
    echo "Error: Directory ${BASE_DIR} already exists"
    exit 1
fi

mkdir -p "${BASE_DIR}/domain"
mkdir -p "${BASE_DIR}/application/command"
mkdir -p "${BASE_DIR}/application/query"
mkdir -p "${BASE_DIR}/adapter/driven/inmemory"
mkdir -p "${BASE_DIR}/adapter/driving/http"
mkdir -p "${BASE_DIR}/adapter/driving/consumer"

# Empty directories are not tracked by git.
touch "${BASE_DIR}/domain/.gitkeep"
touch "${BASE_DIR}/application/command/.gitkeep"
touch "${BASE_DIR}/application/query/.gitkeep"
touch "${BASE_DIR}/adapter/driven/inmemory/.gitkeep"
touch "${BASE_DIR}/adapter/driving/http/.gitkeep"
touch "${BASE_DIR}/adapter/driving/consumer/.gitkeep"

cat > "${BASE_DIR}/AGENTS.md" << TEMPLATE
# ${BC_NAME}

<!-- Two lines: what this bounded context does, and what it leaves to others. -->

## Language

Shared terms: [\`domain.md\`](../../../../documentation/glossary/domain.md), [\`technical.md\`](../../../../documentation/glossary/technical.md).

<!-- Only the terms local to this bounded context, one line each. -->

## Boundaries

<!-- Who calls it; public events consumed and produced. -->

## Flows

<!-- Journeys that cross several use cases; state machines as a compact text diagram. -->

## Invariants and pitfalls

<!-- The rules an agent would break without knowing them. Nothing the code already says. -->
TEMPLATE

echo "Bounded context '${BC_NAME}' created at apps/api/internal/${BC_NAME}"
