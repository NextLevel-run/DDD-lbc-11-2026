import type { AnnonceEnFile } from '#/features/moderation/in-memory-moderation-gateway'

// An annonce waiting in the file de modération, free by default; override what
// the test is about.
export function uneAnnonceEnFile(
  overrides: Partial<AnnonceEnFile> = {},
): AnnonceEnFile {
  return {
    id: 'a1',
    titre: 'Vélo de ville',
    description: 'Très bon état.',
    prix: 15000,
    categorie: 'vehicules',
    ville: 'Lyon',
    codePostal: '69003',
    pseudo: 'cyclo',
    deposeeLe: '2026-11-19T08:00:00Z',
    misEnFileLe: '2026-11-19T09:00:00Z',
    ...overrides,
  }
}
