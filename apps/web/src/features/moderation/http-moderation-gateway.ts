import { ApiError, request } from '#/lib/http'
import type {
  AnnonceAModerer,
  AnnonceAModererResume,
  ModerationGateway,
} from './moderation-gateway'

const chemin = (id: string) => `/moderation/annonces/${encodeURIComponent(id)}`

const corps = (moderateur: string) => JSON.stringify({ moderateur })

export function createHttpModerationGateway(): ModerationGateway {
  return {
    listerAnnoncesAModerer: async () => {
      const { annonces } = await request<{
        annonces: Array<AnnonceAModererResume>
      }>('/moderation/annonces')
      return annonces
    },
    consulterAnnonceAModerer: async (id) => {
      try {
        return await request<AnnonceAModerer>(chemin(id))
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          return null
        }
        throw error
      }
    },
    verrouillerAnnonce: ({ id, moderateur }) =>
      request<void>(`${chemin(id)}/verrou`, {
        method: 'POST',
        body: corps(moderateur),
      }),
    relacherVerrou: ({ id, moderateur }) =>
      request<void>(`${chemin(id)}/relachement`, {
        method: 'POST',
        body: corps(moderateur),
      }),
    validerAnnonce: ({ id, moderateur }) =>
      request<void>(`${chemin(id)}/validation`, {
        method: 'POST',
        body: corps(moderateur),
      }),
    refuserAnnonce: ({ id, moderateur }) =>
      request<void>(`${chemin(id)}/refus`, {
        method: 'POST',
        body: corps(moderateur),
      }),
    challengerAnnonce: ({ id, moderateur, motif }) =>
      request<void>(`${chemin(id)}/challenge`, {
        method: 'POST',
        body: JSON.stringify({ moderateur, motif }),
      }),
  }
}
