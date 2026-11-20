export interface Categorie {
  code: string
  libelle: string
}

export interface Localisation {
  ville: string
  codePostal: string
  departement: { code: string; nom: string }
  region: { code: string; nom: string }
}

export interface Annonce {
  id: string
  titre: string
  description: string
  // In centimes d'euro; zero is a don.
  prix: number
  categorie: Categorie
  localisation: Localisation
  // Only the pseudo: the vendeur's contact details never reach the UI.
  vendeur: { pseudo: string }
  // Image URLs, empty when the annonce has none.
  images: Array<string>
  // ISO 8601 date-time.
  deposeeLe: string
}

// An annonce as a list of results shows it.
export interface AnnonceResume {
  id: string
  titre: string
  // In centimes d'euro; zero is a don.
  prix: number
  ville: string
  images: Array<string>
  // ISO 8601 date-time.
  deposeeLe: string
}

// The criteria of a recherche, all optional; the page counts from 1.
export interface CriteresRecherche {
  motCle?: string
  // A catégorie code.
  categorie?: string
  // Bounds of the fourchette de prix, in centimes, both included.
  prixMin?: number
  prixMax?: number
  // The lieu: an INSEE region or département code, never both.
  region?: string
  departement?: string
  // Five digits are a code postal, anything else a ville.
  villeOuCodePostal?: string
  page?: number
}

// A région and its départements, as the lieu selector shows them.
export interface RegionAvecDepartements {
  code: string
  nom: string
  departements: Array<{ code: string; nom: string }>
}

// One page of results, most recent first, and what a pagination needs.
export interface ResultatRecherche {
  annonces: Array<AnnonceResume>
  total: number
  page: number
  parPage: number
}

// What a vendeur fills in to deposit an annonce, already cleaned up by the
// form: prix in centimes, telephone empty when none.
export interface DeposerAnnonceArgs {
  titre: string
  description: string
  prix: number
  categorie: string
  ville: string
  codePostal: string
  pseudo: string
  email: string
  telephone: string
  // The mot de passe de l'annonce, at least 8 characters; it only goes to the
  // API, which keeps it hashed.
  motDePasse: string
}

// The fields of DeposerAnnonceArgs, as the API names them in `champ`.
export type ChampDepot = keyof DeposerAnnonceArgs

// What a vendeur sends to withdraw an annonce: the mot de passe de l'annonce
// and a code of the closed list of raisons.
export interface SupprimerAnnonceArgs {
  id: string
  motDePasse: string
  raison: string
}

// The fields of the suppression form, as the API names them in `champ`.
export type ChampSuppression = 'motDePasse' | 'raison'

// The fields of a challengee annonce a vendeur may change, as the correction
// form is prefilled with, and the motif of the challenge.
export interface AnnonceAEditer {
  titre: string
  description: string
  // In centimes d'euro; zero is a don.
  prix: number
  // A catégorie code.
  categorie: string
  ville: string
  codePostal: string
  // Why moderation sent the annonce back.
  motif: string
}

// What a vendeur sends to correct a challengee annonce: the mot de passe de
// l'annonce and the fields of a dépôt they may change, already cleaned up by
// the form (prix in centimes).
export interface CorrigerAnnonceArgs {
  id: string
  motDePasse: string
  titre: string
  description: string
  prix: number
  categorie: string
  ville: string
  codePostal: string
}

// Port: what the UI needs from the API about annonces.
export interface AnnonceGateway {
  // Resolves to null when no consultable annonce has this ID.
  consulterAnnonce: (id: string) => Promise<Annonce | null>
  // Only activee annonces; an empty mot-clé finds them all.
  rechercherAnnonces: (
    criteres: CriteresRecherche,
  ) => Promise<ResultatRecherche>
  // Every région, each with its départements: the lieux a recherche takes.
  listerRegions: () => Promise<Array<RegionAvecDepartements>>
  // Resolves to the ID of the new annonce. A refused dépôt rejects with an
  // ApiError whose `champ`, when set, names the field at fault.
  deposerAnnonce: (args: DeposerAnnonceArgs) => Promise<string>
  // Resolves once the annonce is withdrawn. A refusal (wrong mot de passe,
  // unknown raison) rejects with an ApiError whose `champ` names the field;
  // an unknown or already supprimee annonce rejects with a 404.
  supprimerAnnonce: (args: SupprimerAnnonceArgs) => Promise<void>
  // Resolves to what prefills the correction form. A wrong mot de passe
  // rejects with an ApiError whose `champ` is motDePasse; an annonce that is
  // unknown or not challengee rejects with a 404.
  editerAnnonce: (args: {
    id: string
    motDePasse: string
  }) => Promise<AnnonceAEditer>
  // Resolves once the annonce is back in moderation. The refusals are the
  // dépôt's (an ApiError whose `champ` names the field), the wrong mot de
  // passe's and the 404's of editerAnnonce.
  corrigerAnnonce: (args: CorrigerAnnonceArgs) => Promise<void>
}
