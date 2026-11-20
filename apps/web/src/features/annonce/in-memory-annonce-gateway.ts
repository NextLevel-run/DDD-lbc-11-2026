import { ApiError } from '#/lib/http'
import { libelleDeCategorie } from './categories'
import { raisonsDeSuppression } from './raisons-suppression'
import type {
  Annonce,
  AnnonceGateway,
  ChampDepot,
  CriteresRecherche,
  DeposerAnnonceArgs,
  RegionAvecDepartements,
  ResultatRecherche,
} from './annonce-gateway'

const PAR_PAGE = 20

// A few régions of the API's table, enough for the tests: homonyms (Saint-Denis
// 93 and 974), Corse and its two départements.
const regionsParDefaut: Array<RegionAvecDepartements> = [
  {
    code: '84',
    nom: 'Auvergne-Rhône-Alpes',
    departements: [
      { code: '42', nom: 'Loire' },
      { code: '69', nom: 'Rhône' },
    ],
  },
  {
    code: '94',
    nom: 'Corse',
    departements: [
      { code: '2A', nom: 'Corse-du-Sud' },
      { code: '2B', nom: 'Haute-Corse' },
    ],
  },
  {
    code: '11',
    nom: 'Île-de-France',
    departements: [
      { code: '75', nom: 'Paris' },
      { code: '93', nom: 'Seine-Saint-Denis' },
    ],
  },
  {
    code: '04',
    nom: 'La Réunion',
    departements: [{ code: '974', nom: 'La Réunion' }],
  },
]

export function createInMemoryAnnonceGateway(
  options: {
    annonces?: Array<Annonce>
    regions?: Array<RegionAvecDepartements>
    // Makes every dépôt fail as the API would refuse it.
    depotRefuse?: { erreur: string; champ?: ChampDepot }
    // The mot de passe of the annonces given above (default « demo1234 »).
    motDePasse?: string
    // Makes every contact fail as the API does when the email cannot leave.
    envoiImpossible?: boolean
  } = {},
): AnnonceGateway {
  const annonces = new Map(
    (options.annonces ?? []).map((annonce) => [annonce.id, annonce]),
  )
  // The mot de passe of each annonce, kept apart: it never reaches a view.
  const motsDePasse = new Map(
    [...annonces.keys()].map((id) => [id, options.motDePasse ?? 'demo1234']),
  )
  return {
    consulterAnnonce: async (id) => annonces.get(id) ?? null,
    rechercherAnnonces: async (criteres) =>
      rechercher([...annonces.values()], criteres),
    listerRegions: async () => options.regions ?? regionsParDefaut,
    deposerAnnonce: async (args) => {
      if (options.depotRefuse) {
        const { erreur, champ } = options.depotRefuse
        throw new ApiError(erreur, 400, champ)
      }
      // Mirrors the API's rule on the mot de passe de l'annonce.
      if ([...args.motDePasse].length < 8) {
        throw new ApiError(
          "Le mot de passe de l'annonce doit comporter au moins 8 caractères",
          400,
          'motDePasse',
        )
      }
      const annonce = toAnnonce(crypto.randomUUID(), args)
      annonces.set(annonce.id, annonce)
      motsDePasse.set(annonce.id, args.motDePasse)
      return annonce.id
    },
    // Mirrors the API: the raison is checked first, then the annonce, then
    // the mot de passe; a withdrawn annonce is gone, as a 404 would say.
    supprimerAnnonce: async ({ id, motDePasse, raison }) => {
      if (!raisonsDeSuppression.some(({ code }) => code === raison)) {
        throw new ApiError(
          'Choisissez une raison de suppression',
          400,
          'raison',
        )
      }
      if (!annonces.has(id)) {
        throw new ApiError('Annonce introuvable', 404)
      }
      if (motsDePasse.get(id) !== motDePasse) {
        throw new ApiError(
          "Le mot de passe de l'annonce est incorrect",
          400,
          'motDePasse',
        )
      }
      annonces.delete(id)
      motsDePasse.delete(id)
    },
    // The form checks the fields: only the annonce and the email are left.
    contacter: async ({ id }) => {
      if (!annonces.has(id)) {
        throw new ApiError('Annonce introuvable', 404)
      }
      if (options.envoiImpossible) {
        throw new ApiError(
          "Votre message n'a pas pu être envoyé, réessayez plus tard.",
          502,
        )
      }
    },
  }
}

// The in-memory adapter knows no département table: it keeps the code postal
// prefix as the département code and leaves the names empty.
function toAnnonce(id: string, args: DeposerAnnonceArgs): Annonce {
  return {
    id,
    titre: args.titre,
    description: args.description,
    prix: args.prix,
    categorie: {
      code: args.categorie,
      libelle: libelleDeCategorie(args.categorie) ?? args.categorie,
    },
    localisation: {
      ville: args.ville,
      codePostal: args.codePostal,
      departement: { code: args.codePostal.slice(0, 2), nom: '' },
      region: { code: '', nom: '' },
    },
    vendeur: { pseudo: args.pseudo },
    images: [],
    deposeeLe: new Date().toISOString(),
  }
}

// Lower case, no accent, hyphens and apostrophes as spaces, runs of spaces
// collapsed: what the API compares the mot-clé and the ville on.
const normaliser = (texte: string) =>
  texte
    .toLowerCase()
    .replace(/œ/g, 'oe')
    .replace(/æ/g, 'ae')
    .normalize('NFD')
    .replace(/\p{M}/gu, '')
    .replace(/[\s'’-]+/g, ' ')
    .trim()

// Mirrors the API: five digits are an exact code postal, anything else a
// ville identical once normalised.
function retientVilleOuCodePostal(annonce: Annonce, saisie = ''): boolean {
  const texte = saisie.trim()
  if (texte === '') {
    return true
  }
  if (/^\d{5}$/.test(texte)) {
    return annonce.localisation.codePostal === texte
  }
  return normaliser(annonce.localisation.ville) === normaliser(texte)
}

// Mirrors the API: every word of the mot-clé in the titre or the
// description, the catégorie, the prix within both bounds (included), the
// région or département and the ville or code postal, most recent first,
// PAR_PAGE a page. The annonces given to the in-memory adapter
// are all activee.
function rechercher(
  annonces: Array<Annonce>,
  {
    motCle = '',
    categorie,
    prixMin,
    prixMax,
    region,
    departement,
    villeOuCodePostal,
    page = 1,
  }: CriteresRecherche,
): ResultatRecherche {
  const mots = normaliser(motCle).split(' ').filter(Boolean)
  const retenues = annonces
    .filter(
      (annonce) =>
        (categorie === undefined || annonce.categorie.code === categorie) &&
        (prixMin === undefined || annonce.prix >= prixMin) &&
        (prixMax === undefined || annonce.prix <= prixMax) &&
        (region === undefined || annonce.localisation.region.code === region) &&
        (departement === undefined ||
          annonce.localisation.departement.code === departement) &&
        retientVilleOuCodePostal(annonce, villeOuCodePostal),
    )
    .filter((annonce) => {
      const texte = normaliser(`${annonce.titre} ${annonce.description}`)
      return mots.every((mot) => texte.includes(mot))
    })
    .sort((a, b) => b.deposeeLe.localeCompare(a.deposeeLe))
  const start = (page - 1) * PAR_PAGE
  return {
    annonces: retenues
      .slice(start, start + PAR_PAGE)
      .map(({ id, titre, prix, localisation, images, deposeeLe }) => ({
        id,
        titre,
        prix,
        ville: localisation.ville,
        images,
        deposeeLe,
      })),
    total: retenues.length,
    page,
    parPage: PAR_PAGE,
  }
}
