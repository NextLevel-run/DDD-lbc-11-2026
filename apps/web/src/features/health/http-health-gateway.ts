import { request } from '#/lib/http'
import type { Health, HealthGateway } from './health-gateway'

export function createHttpHealthGateway(): HealthGateway {
  return {
    check: () => request<Health>('/health'),
  }
}
