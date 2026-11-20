// An annonce of the file de modération, as the list shows it.
export interface AnnonceAModererResume {
  id: string
  titre: string
  // In centimes d'euro; zero is a don.
  prix: number
  ville: string
  pseudo: string
  // ISO 8601 date-time.
  misEnFileLe: string
  // The moderator holding the lock, empty when the annonce is free.
  verrouillePar: string
  verrouilleDepuis?: string
}

// A dated line of the history of an annonce.
export interface EvenementHistorique {
  // The API's stable event type: AnnonceMiseEnFile, AnnonceVerrouillee,
  // VerrouAnnonceRelache…
  type: string
  date: string
  // Empty when no moderator is behind the event (a lock that expired).
  moderateur: string
  motif: string
}

// `en-file` and `verrouillee` wait for a moderator; the others are decided and
// out of the file.
export type StatutModeration =
  'en-file' | 'verrouillee' | 'validee' | 'refusee' | 'challengee'

// An annonce to moderate: what the deposit sent, the lock and the history.
export interface AnnonceAModerer {
  id: string
  titre: string
  description: string
  prix: number
  // A catégorie code.
  categorie: string
  ville: string
  codePostal: string
  pseudo: string
  deposeeLe: string
  statut: StatutModeration
  verrouillePar: string
  verrouilleDepuis?: string
  historique: Array<EvenementHistorique>
}

// What a moderator sends to lock or release an annonce.
export interface ModerateurArgs {
  id: string
  // The pseudo of the moderator, 2 to 30 characters.
  moderateur: string
}

// What a moderator sends to challenge an annonce.
export interface ChallengerArgs extends ModerateurArgs {
  // Why the vendeur must correct the annonce, 10 to 1 000 characters.
  motif: string
}

// Port: what the UI needs from the API about moderation. A refused action
// rejects with an ApiError: 400 on `moderateur`, 404 for an unknown annonce,
// 409 for a lock or a state that forbids it; a decision also needs the valid
// lock of the moderator, and challenging a valid `motif` (400 on `motif`).
export interface ModerationGateway {
  // The annonces waiting for a moderator, oldest first.
  listerAnnoncesAModerer: () => Promise<Array<AnnonceAModererResume>>
  // Resolves to null when the annonce is unknown.
  consulterAnnonceAModerer: (id: string) => Promise<AnnonceAModerer | null>
  verrouillerAnnonce: (args: ModerateurArgs) => Promise<void>
  relacherVerrou: (args: ModerateurArgs) => Promise<void>
  // Puts the annonce online; its vendeur is told by email.
  validerAnnonce: (args: ModerateurArgs) => Promise<void>
  // Keeps the annonce offline for good; no email (fraud).
  refuserAnnonce: (args: ModerateurArgs) => Promise<void>
  // Sends the annonce back to its vendeur, with the motif.
  challengerAnnonce: (args: ChallengerArgs) => Promise<void>
}
