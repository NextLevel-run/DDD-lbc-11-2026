import { mutationOptions, queryOptions } from '@tanstack/react-query'
import type { QueryClient } from '@tanstack/react-query'
import type {
  ChallengerArgs,
  ModerateurArgs,
  ModerationGateway,
} from './moderation-gateway'

export const annoncesAModererQueryOptions = (gateway: ModerationGateway) =>
  queryOptions({
    queryKey: ['moderation', 'annonces'],
    queryFn: () => gateway.listerAnnoncesAModerer(),
  })

export const annonceAModererQueryOptions = (
  gateway: ModerationGateway,
  id: string,
) =>
  queryOptions({
    queryKey: ['moderation', 'annonces', id],
    queryFn: () => gateway.consulterAnnonceAModerer(id),
  })

// A lock changes the file and the detail: every read of moderation is stale.
export const verrouillerAnnonceMutationOptions = (
  gateway: ModerationGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: ModerateurArgs) => gateway.verrouillerAnnonce(args),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['moderation'] }),
  })

export const relacherVerrouMutationOptions = (
  gateway: ModerationGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: ModerateurArgs) => gateway.relacherVerrou(args),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['moderation'] }),
  })

// A decision takes the annonce out of the file: every read is stale.
export const validerAnnonceMutationOptions = (
  gateway: ModerationGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: ModerateurArgs) => gateway.validerAnnonce(args),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['moderation'] }),
  })

export const refuserAnnonceMutationOptions = (
  gateway: ModerationGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: ModerateurArgs) => gateway.refuserAnnonce(args),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['moderation'] }),
  })

export const challengerAnnonceMutationOptions = (
  gateway: ModerationGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: ChallengerArgs) => gateway.challengerAnnonce(args),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['moderation'] }),
  })
