import { mutationOptions, queryOptions } from '@tanstack/react-query'
import type { QueryClient } from '@tanstack/react-query'
import type {
  AnnonceGateway,
  CorrigerAnnonceArgs,
  CriteresRecherche,
  DeposerAnnonceArgs,
  SupprimerAnnonceArgs,
} from './annonce-gateway'

export const annonceQueryOptions = (gateway: AnnonceGateway, id: string) =>
  queryOptions({
    queryKey: ['annonces', id],
    queryFn: () => gateway.consulterAnnonce(id),
  })

export const rechercherAnnoncesQueryOptions = (
  gateway: AnnonceGateway,
  criteres: CriteresRecherche,
) =>
  queryOptions({
    queryKey: ['annonces', 'recherche', criteres],
    queryFn: () => gateway.rechercherAnnonces(criteres),
  })

// The régions never change while the app runs.
export const regionsQueryOptions = (gateway: AnnonceGateway) =>
  queryOptions({
    queryKey: ['regions'],
    queryFn: () => gateway.listerRegions(),
    staleTime: Infinity,
  })

// A new annonce makes every read of annonces stale.
export const deposerAnnonceMutationOptions = (
  gateway: AnnonceGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: DeposerAnnonceArgs) => gateway.deposerAnnonce(args),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['annonces'] }),
  })

// A withdrawn annonce makes every read of annonces stale.
export const supprimerAnnonceMutationOptions = (
  gateway: AnnonceGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: SupprimerAnnonceArgs) => gateway.supprimerAnnonce(args),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['annonces'] }),
  })

// Opening the correction is a write of its own kind: the mot de passe is sent
// and the answer is what prefills the form, kept out of the query cache.
export const editerAnnonceMutationOptions = (gateway: AnnonceGateway) =>
  mutationOptions({
    mutationFn: (args: { id: string; motDePasse: string }) =>
      gateway.editerAnnonce(args),
  })

// A corrected annonce leaves the public reads: every one of them is stale.
export const corrigerAnnonceMutationOptions = (
  gateway: AnnonceGateway,
  queryClient: QueryClient,
) =>
  mutationOptions({
    mutationFn: (args: CorrigerAnnonceArgs) => gateway.corrigerAnnonce(args),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['annonces'] }),
  })
