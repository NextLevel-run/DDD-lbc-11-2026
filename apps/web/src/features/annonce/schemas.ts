import { z } from 'zod'
import { categories } from './categories'
import { raisonsDeSuppression } from './raisons-suppression'
import type {
  CriteresRecherche,
  DeposerAnnonceArgs,
  RegionAvecDepartements,
} from './annonce-gateway'

// The form values of the dépôt: every field is the text the vendeur typed.
export type DeposerAnnonceValues = Record<keyof DeposerAnnonceArgs, string>

// Reads a prix typed in euros, with a comma or a dot and up to two decimals,
// as centimes. Undefined when the text is no such amount.
export function eurosEnCentimes(euros: string): number | undefined {
  const match = /^(\d+)(?:[.,](\d{1,2}))?$/.exec(euros.replace(/\s/g, ''))
  if (!match) {
    return undefined
  }
  const [, entiers, decimales = ''] = match
  return Number(entiers) * 100 + Number(decimales.padEnd(2, '0'))
}

// Mirrors the API's rule: 01-95, Corse 200-206, overseas 971-974 and 976.
function designeUnDepartement(codePostal: string): boolean {
  const prefixe = Number(codePostal.slice(0, 2))
  if (prefixe === 20) {
    return Number(codePostal[2]) <= 6
  }
  if (prefixe === 97) {
    return ['1', '2', '3', '4', '6'].includes(codePostal[2])
  }
  return prefixe >= 1 && prefixe <= 95
}

const longueur = (min: number, max: number, message: string) =>
  z.string().trim().min(min, message).max(max, message)

// The vendeur's and the acheteur's share these rules, as in the API.
const pseudo = longueur(2, 30, 'Le pseudo doit comporter de 2 à 30 caractères')
const email = z
  .string()
  .trim()
  .regex(/^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$/, "L'email n'est pas valide")

// The rules of the domain, checked in the browser for an immediate feedback.
// The API has the final say: its errors are shown too. Parsing turns the
// values into DeposerAnnonceArgs.
export const deposerAnnonceSchema = z.object({
  titre: longueur(3, 100, 'Le titre doit comporter de 3 à 100 caractères'),
  description: longueur(
    10,
    4000,
    'La description doit comporter de 10 à 4 000 caractères',
  ),
  prix: z
    .string()
    .trim()
    .min(1, 'Le prix est obligatoire')
    .transform((euros, ctx) => {
      const centimes = eurosEnCentimes(euros)
      if (centimes === undefined) {
        ctx.addIssue({
          code: 'custom',
          message: 'Le prix doit être un montant en euros, par exemple 12,50',
        })
        return z.NEVER
      }
      return centimes
    }),
  categorie: z
    .string()
    .refine((code) => categories.some((categorie) => categorie.code === code), {
      message: 'La catégorie est obligatoire',
    }),
  ville: z.string().trim().min(1, 'La ville est obligatoire'),
  codePostal: z
    .string()
    .trim()
    .regex(/^\d{5}$/, {
      message: 'Le code postal doit comporter 5 chiffres',
      abort: true,
    })
    .refine(designeUnDepartement, {
      message: 'Le code postal ne désigne aucun département',
    }),
  pseudo,
  email,
  telephone: z
    .string()
    .transform((telephone) => telephone.replace(/[\s.-]/g, ''))
    .refine((telephone) => telephone === '' || /^0\d{9}$/.test(telephone), {
      message: 'Le téléphone doit comporter 10 chiffres et commencer par 0',
    }),
  motDePasse: z
    .string()
    .min(
      8,
      "Le mot de passe de l'annonce doit comporter au moins 8 caractères",
    ),
}) satisfies z.ZodType<DeposerAnnonceArgs, DeposerAnnonceValues>

// The suppression form: the raison and the mot de passe de l'annonce. The
// API checks the mot de passe; the form only asks for one.
export const supprimerAnnonceSchema = z.object({
  raison: z
    .string()
    .refine((code) => raisonsDeSuppression.some((r) => r.code === code), {
      message: 'Choisissez une raison',
    }),
  motDePasse: z.string().min(1, "Saisissez le mot de passe de l'annonce"),
})

export type SupprimerAnnonceValues = z.infer<typeof supprimerAnnonceSchema>

// The contact form: the acheteur's pseudo and email, and their message,
// 10 to 2 000 characters once trimmed.
export const contacterSchema = z.object({
  pseudo,
  email,
  message: longueur(
    10,
    2000,
    'Le message doit comporter de 10 à 2 000 caractères',
  ),
})

export type ContacterValues = z.input<typeof contacterSchema>

// A text read from the URL (mot-clé, code, ville): a number as text, a blank
// as none.
function lireTexte(value: unknown): unknown {
  if (typeof value === 'number') {
    return String(value)
  }
  if (typeof value === 'string' && value.trim() === '') {
    return undefined
  }
  return value
}

// A prix bound read from the URL, in euros: text is read as a number, a
// blank as none.
function lirePrix(value: unknown): unknown {
  if (typeof value === 'string') {
    return value.trim() === '' ? undefined : Number(value)
  }
  return value
}

const codesDeCategorie = categories.map((categorie) => categorie.code) as [
  string,
  ...Array<string>,
]

const texte = z
  .preprocess(lireTexte, z.string().trim().min(1).optional())
  .catch(undefined)

// An INSEE code as the URL carries it: two or three digits, or 2A and 2B.
const codeInsee = z
  .preprocess(
    lireTexte,
    z
      .string()
      .regex(/^(\d{2,3}|2[AB])$/)
      .optional(),
  )
  .catch(undefined)

// A bound past the safe integers once in centimes could not reach the API
// intact: it is dropped too.
const borneDePrix = z
  .preprocess(
    lirePrix,
    z
      .number()
      .finite()
      .min(0)
      .max(Number.MAX_SAFE_INTEGER / 100)
      .optional(),
  )
  .catch(undefined)

// The search params of /recherche. The router parses `?motCle=123` as a
// number, so the texts are read back as text; the prix bounds are in euros.
// Anything unusable is dropped rather than failing the page, a maximum below
// the minimum included, and the département when a région is set too.
export const rechercheSearchSchema = z
  .object({
    motCle: texte,
    categorie: z.enum(codesDeCategorie).optional().catch(undefined),
    prixMin: borneDePrix,
    prixMax: borneDePrix,
    region: codeInsee,
    departement: codeInsee,
    villeOuCodePostal: texte,
    page: z.coerce.number().int().min(1).optional().catch(undefined),
  })
  .transform((search) =>
    search.prixMin !== undefined &&
    search.prixMax !== undefined &&
    search.prixMax < search.prixMin
      ? { ...search, prixMax: undefined }
      : search,
  )
  .transform((search) =>
    search.region !== undefined && search.departement !== undefined
      ? { ...search, departement: undefined }
      : search,
  )

export type RechercheSearch = z.infer<typeof rechercheSearchSchema>

const enCentimes = (euros: number | undefined) =>
  euros === undefined ? undefined : Math.round(euros * 100)

// The criteria the gateway takes, from the search params of the URL: the prix
// bounds go from euros to centimes, and a région or département missing from
// `regions` is dropped, as the API would refuse it.
export function criteresDeRecherche(
  search: Partial<RechercheSearch>,
  regions: Array<RegionAvecDepartements> = [],
): CriteresRecherche {
  const regionConnue = regions.some(({ code }) => code === search.region)
  const departementConnu = regions.some(({ departements }) =>
    departements.some(({ code }) => code === search.departement),
  )
  const criteres: CriteresRecherche = {
    motCle: search.motCle,
    categorie: search.categorie,
    prixMin: enCentimes(search.prixMin),
    prixMax: enCentimes(search.prixMax),
    region: regionConnue ? search.region : undefined,
    departement: departementConnu ? search.departement : undefined,
    villeOuCodePostal: search.villeOuCodePostal,
    page: search.page,
  }
  return Object.fromEntries(
    Object.entries(criteres).filter(([, value]) => value !== undefined),
  )
}
