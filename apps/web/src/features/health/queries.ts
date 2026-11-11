import { queryOptions } from '@tanstack/react-query'
import type { HealthGateway } from './health-gateway'

export const healthQueryOptions = (gateway: HealthGateway) =>
  queryOptions({
    queryKey: ['health'],
    queryFn: () => gateway.check(),
  })
