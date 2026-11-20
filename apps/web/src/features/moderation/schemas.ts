import { z } from 'zod'

// Mirrors the API's rule: the pseudo of a moderator holds 2 to 30 characters
// once trimmed. The API has the final say.
export const moderateurSchema = z.object({
  moderateur: z
    .string()
    .trim()
    .min(2, 'Le pseudo doit comporter de 2 à 30 caractères')
    .max(30, 'Le pseudo doit comporter de 2 à 30 caractères'),
})

// Mirrors the API's rule: the motif of a challenge holds 10 to 1 000
// characters once trimmed. The API has the final say.
export const motifSchema = z.object({
  motif: z
    .string()
    .trim()
    .min(10, 'Le motif doit comporter de 10 à 1 000 caractères')
    .max(1000, 'Le motif doit comporter de 10 à 1 000 caractères'),
})
