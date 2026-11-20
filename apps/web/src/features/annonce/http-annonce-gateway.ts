import { ApiError, request } from '#/lib/http'
import type {
  Annonce,
  AnnonceAEditer,
  AnnonceGateway,
  RegionAvecDepartements,
  ResultatRecherche,
} from './annonce-gateway'

export function createHttpAnnonceGateway(): AnnonceGateway {
  return {
    consulterAnnonce: async (id) => {
      try {
        return await request<Annonce>(`/annonces/${encodeURIComponent(id)}`)
      } catch (error) {
        // 404: unknown or not consultable; 400: not an annonce ID at all.
        if (
          error instanceof ApiError &&
          (error.status === 404 || error.status === 400)
        ) {
          return null
        }
        throw error
      }
    },
    rechercherAnnonces: async ({
      motCle,
      categorie,
      prixMin,
      prixMax,
      region,
      departement,
      villeOuCodePostal,
      page,
    }) => {
      const params = new URLSearchParams()
      if (motCle) {
        params.set('motCle', motCle)
      }
      if (categorie) {
        params.set('categorie', categorie)
      }
      // Zero is a bound too: a fourchette from 0 to 0 finds the dons.
      if (prixMin !== undefined) {
        params.set('prixMin', String(prixMin))
      }
      if (prixMax !== undefined) {
        params.set('prixMax', String(prixMax))
      }
      if (region) {
        params.set('region', region)
      }
      if (departement) {
        params.set('departement', departement)
      }
      if (villeOuCodePostal) {
        params.set('villeOuCodePostal', villeOuCodePostal)
      }
      if (page && page > 1) {
        params.set('page', String(page))
      }
      const query = params.toString()
      return request<ResultatRecherche>(`/annonces${query ? `?${query}` : ''}`)
    },
    listerRegions: () => request<Array<RegionAvecDepartements>>('/regions'),
    deposerAnnonce: async (args) => {
      const { id } = await request<{ id: string }>('/annonces', {
        method: 'POST',
        body: JSON.stringify(args),
      })
      return id
    },
    supprimerAnnonce: async ({ id, motDePasse, raison }) => {
      await request<void>(`/annonces/${encodeURIComponent(id)}/suppression`, {
        method: 'POST',
        body: JSON.stringify({ motDePasse, raison }),
      })
    },
    editerAnnonce: ({ id, motDePasse }) =>
      request<AnnonceAEditer>(`/annonces/${encodeURIComponent(id)}/edition`, {
        method: 'POST',
        body: JSON.stringify({ motDePasse }),
      }),
    corrigerAnnonce: async ({ id, ...correction }) => {
      await request<void>(`/annonces/${encodeURIComponent(id)}/correction`, {
        method: 'POST',
        body: JSON.stringify(correction),
      })
    },
  }
}
