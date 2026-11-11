import type { HealthGateway } from './health-gateway'

export function createInMemoryHealthGateway(
  options: { reachable?: boolean } = {},
): HealthGateway {
  const { reachable = true } = options
  return {
    check: async () => {
      if (!reachable) {
        throw new Error('api unreachable')
      }
      return { status: 'ok' }
    },
  }
}
