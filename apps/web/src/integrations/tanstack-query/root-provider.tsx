import { QueryClient } from '@tanstack/react-query'
import { createHttpGateways } from '#/gateways'
import type { Gateways } from '#/gateways'

export interface RouterContext {
  queryClient: QueryClient
  gateways: Gateways
}

export function getContext(): RouterContext {
  return {
    queryClient: new QueryClient(),
    gateways: createHttpGateways(),
  }
}
