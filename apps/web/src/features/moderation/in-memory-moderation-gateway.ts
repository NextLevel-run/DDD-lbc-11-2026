import { ApiError } from '#/lib/http'
import type {
  AnnonceAModerer,
  AnnonceAModererResume,
  ModerationGateway,
  StatutModeration,
} from './moderation-gateway'

const DEUX_HEURES = 2 * 60 * 60 * 1000

// An annonce of the file as a test seeds it: the snapshot, when it entered the
// file and, optionally, who holds it since when.
export type AnnonceEnFile = Omit<
  AnnonceAModerer,
  'statut' | 'verrouillePar' | 'historique'
> & {
  misEnFileLe: string
  verrouillePar?: string
  verrouilleDepuis?: string
}

// Mirrors the API's moderation rules: the file is oldest first, a lock holds
// for two hours, another moderator's valid lock refuses with a 409, a decision
// needs the valid lock of its moderator and takes the annonce out of the file.
export function createInMemoryModerationGateway(
  options: {
    annonces?: Array<AnnonceEnFile>
    now?: () => Date
  } = {},
): ModerationGateway {
  const now = options.now ?? (() => new Date())
  const annonces = new Map(
    (options.annonces ?? []).map((annonce) => [
      annonce.id,
      {
        ...annonce,
        decision: undefined as StatutModeration | undefined,
        historique: [
          {
            type: 'AnnonceMiseEnFile',
            date: annonce.misEnFileLe,
            moderateur: '',
            motif: '',
          },
          ...(annonce.verrouillePar
            ? [
                {
                  type: 'AnnonceVerrouillee',
                  date: annonce.verrouilleDepuis ?? annonce.misEnFileLe,
                  moderateur: annonce.verrouillePar,
                  motif: '',
                },
              ]
            : []),
        ],
      },
    ]),
  )

  // The holder of a lock still valid, if any.
  const verrouValide = (id: string) => {
    const annonce = annonces.get(id)
    if (!annonce?.verrouillePar || !annonce.verrouilleDepuis) {
      return undefined
    }
    const expire =
      now().getTime() - new Date(annonce.verrouilleDepuis).getTime() >
      DEUX_HEURES
    return expire
      ? undefined
      : { moderateur: annonce.verrouillePar, depuis: annonce.verrouilleDepuis }
  }

  const exigerModerateur = (moderateur: string) => {
    const longueur = [...moderateur.trim()].length
    if (longueur < 2 || longueur > 30) {
      throw new ApiError(
        'Le pseudo du modérateur doit comporter de 2 à 30 caractères',
        400,
        'moderateur',
      )
    }
  }

  // The annonce, once its moderator holds a valid lock: what every decision
  // and the release need.
  const exigerVerrou = (id: string, moderateur: string) => {
    exigerModerateur(moderateur)
    const annonce = exigerAnnonce(id)
    const pseudo = moderateur.trim()
    if (annonce.decision || annonce.verrouillePar !== pseudo) {
      throw new ApiError('Cette annonce n’est pas verrouillée par vous', 409)
    }
    if (!verrouValide(id)) {
      throw new ApiError('Votre verrou a expiré', 409)
    }
    return { annonce, pseudo }
  }

  const decider = (
    id: string,
    moderateur: string,
    decision: StatutModeration,
    type: string,
    motif = '',
  ) => {
    const { annonce, pseudo } = exigerVerrou(id, moderateur)
    annonce.decision = decision
    annonce.verrouillePar = ''
    annonce.verrouilleDepuis = undefined
    annonce.historique.push({
      type,
      date: now().toISOString(),
      moderateur: pseudo,
      motif,
    })
  }

  const exigerAnnonce = (id: string) => {
    const annonce = annonces.get(id)
    if (!annonce) {
      throw new ApiError('Annonce introuvable', 404)
    }
    return annonce
  }

  return {
    listerAnnoncesAModerer: async () =>
      [...annonces.values()]
        .filter((annonce) => !annonce.decision)
        .sort(
          (a, b) =>
            a.misEnFileLe.localeCompare(b.misEnFileLe) ||
            a.id.localeCompare(b.id),
        )
        .map((annonce): AnnonceAModererResume => ({
          id: annonce.id,
          titre: annonce.titre,
          prix: annonce.prix,
          ville: annonce.ville,
          pseudo: annonce.pseudo,
          misEnFileLe: annonce.misEnFileLe,
          verrouillePar: verrouValide(annonce.id)?.moderateur ?? '',
          verrouilleDepuis: verrouValide(annonce.id)?.depuis,
        })),
    consulterAnnonceAModerer: async (id) => {
      const annonce = annonces.get(id)
      if (!annonce) {
        return null
      }
      const verrou = verrouValide(id)
      return {
        ...annonce,
        statut: annonce.decision ?? (verrou ? 'verrouillee' : 'en-file'),
        verrouillePar: verrou?.moderateur ?? '',
        verrouilleDepuis: verrou?.depuis,
      }
    },
    verrouillerAnnonce: async ({ id, moderateur }) => {
      exigerModerateur(moderateur)
      const annonce = exigerAnnonce(id)
      const pseudo = moderateur.trim()
      if (annonce.decision) {
        throw new ApiError(
          'Cette annonce n’est pas dans la file de modération',
          409,
        )
      }
      const verrou = verrouValide(id)
      if (verrou?.moderateur === pseudo) {
        return
      }
      if (verrou) {
        throw new ApiError(
          'Cette annonce est déjà verrouillée par un autre modérateur',
          409,
        )
      }
      const date = now().toISOString()
      if (annonce.verrouillePar) {
        annonce.historique.push({
          type: 'VerrouAnnonceRelache',
          date,
          moderateur: '',
          motif: '',
        })
      }
      annonce.verrouillePar = pseudo
      annonce.verrouilleDepuis = date
      annonce.historique.push({
        type: 'AnnonceVerrouillee',
        date,
        moderateur: pseudo,
        motif: '',
      })
    },
    relacherVerrou: async ({ id, moderateur }) => {
      const { annonce, pseudo } = exigerVerrou(id, moderateur)
      annonce.verrouillePar = ''
      annonce.verrouilleDepuis = undefined
      annonce.historique.push({
        type: 'VerrouAnnonceRelache',
        date: now().toISOString(),
        moderateur: pseudo,
        motif: '',
      })
    },
    validerAnnonce: async ({ id, moderateur }) =>
      decider(id, moderateur, 'validee', 'AnnonceValidee'),
    refuserAnnonce: async ({ id, moderateur }) =>
      decider(id, moderateur, 'refusee', 'AnnonceRefusee'),
    challengerAnnonce: async ({ id, moderateur, motif }) => {
      exigerModerateur(moderateur)
      const texte = motif.trim()
      const longueur = [...texte].length
      if (longueur < 10 || longueur > 1000) {
        throw new ApiError(
          'Le motif doit comporter de 10 à 1 000 caractères',
          400,
          'motif',
        )
      }
      decider(id, moderateur, 'challengee', 'AnnonceChallengee', texte)
    },
  }
}
