import { createHttpAnnonceGateway } from '#/features/annonce/http-annonce-gateway'
import { createInMemoryAnnonceGateway } from '#/features/annonce/in-memory-annonce-gateway'
import { createHttpModerationGateway } from '#/features/moderation/http-moderation-gateway'
import { createInMemoryModerationGateway } from '#/features/moderation/in-memory-moderation-gateway'
import { createHttpHealthGateway } from '#/features/health/http-health-gateway'
import { createInMemoryHealthGateway } from '#/features/health/in-memory-health-gateway'
import type { AnnonceGateway } from '#/features/annonce/annonce-gateway'
import type { HealthGateway } from '#/features/health/health-gateway'
import type { ModerationGateway } from '#/features/moderation/moderation-gateway'

// Every port the UI depends on. Routes read it from the router context and
// never import an adapter.
export interface Gateways {
  annonce: AnnonceGateway
  health: HealthGateway
  moderation: ModerationGateway
}

export function createHttpGateways(): Gateways {
  return {
    annonce: createHttpAnnonceGateway(),
    health: createHttpHealthGateway(),
    moderation: createHttpModerationGateway(),
  }
}

export function createInMemoryGateways(
  overrides: Partial<Gateways> = {},
): Gateways {
  return {
    annonce: createInMemoryAnnonceGateway(),
    health: createInMemoryHealthGateway(),
    moderation: createInMemoryModerationGateway(),
    ...overrides,
  }
}
