import { createHttpHealthGateway } from '#/features/health/http-health-gateway'
import { createInMemoryHealthGateway } from '#/features/health/in-memory-health-gateway'
import type { HealthGateway } from '#/features/health/health-gateway'

// Every port the UI depends on. Routes read it from the router context and
// never import an adapter.
export interface Gateways {
  health: HealthGateway
}

export function createHttpGateways(): Gateways {
  return {
    health: createHttpHealthGateway(),
  }
}

export function createInMemoryGateways(
  overrides: Partial<Gateways> = {},
): Gateways {
  return {
    health: createInMemoryHealthGateway(),
    ...overrides,
  }
}
