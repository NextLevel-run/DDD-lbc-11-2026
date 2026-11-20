import { describe, expect, it } from 'vitest'
import { ApiError } from '#/lib/http'
import { uneAnnonce } from '#/test/annonces'
import { createInMemoryAnnonceGateway } from './in-memory-annonce-gateway'
import type { DeposerAnnonceArgs } from './annonce-gateway'

const depot: DeposerAnnonceArgs = {
  titre: 'Vélo de course',
  description: 'Cadre carbone, très bon état.',
  prix: 45000,
  categorie: 'loisirs',
  ville: 'Lyon',
  codePostal: '69003',
  pseudo: 'cycliste69',
  email: 'cycliste@example.com',
  telephone: '',
  motDePasse: 'demo1234',
}

describe('in-memory annonce gateway, dépôt', () => {
  it('keeps a deposited annonce out of the consultation and the search', async () => {
    const gateway = createInMemoryAnnonceGateway()

    const id = await gateway.deposerAnnonce(depot)

    expect(id).not.toBe('')
    expect(await gateway.consulterAnnonce(id)).toBeNull()
    expect((await gateway.rechercherAnnonces({})).total).toBe(0)
  })

  it('refuses a mot de passe that is missing or too short, naming the field', async () => {
    const gateway = createInMemoryAnnonceGateway()

    for (const motDePasse of ['', '1234567']) {
      const refus = await gateway
        .deposerAnnonce({ ...depot, motDePasse })
        .catch((error: unknown) => error)

      expect(refus).toBeInstanceOf(ApiError)
      expect((refus as ApiError).champ).toBe('motDePasse')
      expect((refus as ApiError).status).toBe(400)
    }
  })
})

describe('in-memory annonce gateway, suppression', () => {
  const supprimer = (
    gateway: ReturnType<typeof createInMemoryAnnonceGateway>,
    id: string,
    overrides: Partial<{ motDePasse: string; raison: string }> = {},
  ) =>
    gateway.supprimerAnnonce({
      id,
      motDePasse: 'demo1234',
      raison: 'autre',
      ...overrides,
    })

  it('withdraws an annonce with its mot de passe', async () => {
    const annonce = uneAnnonce()
    const gateway = createInMemoryAnnonceGateway({ annonces: [annonce] })
    const id = annonce.id

    await supprimer(gateway, id)

    expect(await gateway.consulterAnnonce(id)).toBeNull()
    expect((await gateway.rechercherAnnonces({})).total).toBe(0)
  })

  it('refuses a wrong mot de passe and keeps the annonce', async () => {
    const annonce = uneAnnonce()
    const gateway = createInMemoryAnnonceGateway({ annonces: [annonce] })
    const id = annonce.id

    const refus = await supprimer(gateway, id, { motDePasse: 'mauvais' }).catch(
      (error: unknown) => error,
    )

    expect((refus as ApiError).champ).toBe('motDePasse')
    expect(await gateway.consulterAnnonce(id)).not.toBeNull()
  })

  it('refuses an unknown raison, naming the field', async () => {
    const annonce = uneAnnonce()
    const gateway = createInMemoryAnnonceGateway({ annonces: [annonce] })
    const id = annonce.id

    const refus = await supprimer(gateway, id, { raison: '' }).catch(
      (error: unknown) => error,
    )

    expect((refus as ApiError).champ).toBe('raison')
  })

  it('answers 404 for an unknown or already withdrawn annonce', async () => {
    const annonce = uneAnnonce()
    const gateway = createInMemoryAnnonceGateway({ annonces: [annonce] })
    const id = annonce.id
    await supprimer(gateway, id)

    const refus = await supprimer(gateway, id).catch((error: unknown) => error)

    expect((refus as ApiError).status).toBe(404)
  })

  it('answers 404 for an annonce still awaiting moderation', async () => {
    const gateway = createInMemoryAnnonceGateway()
    const id = await gateway.deposerAnnonce(depot)

    const refus = await supprimer(gateway, id).catch((error: unknown) => error)

    expect((refus as ApiError).status).toBe(404)
  })
})

describe('in-memory annonce gateway, correction', () => {
  const challengee = {
    id: 'a-corriger',
    titre: 'Vélo de course',
    description: 'Cadre carbone, très bon état.',
    prix: 45000,
    categorie: 'loisirs',
    ville: 'Lyon',
    codePostal: '69003',
    motif: 'Photos manquantes',
  }
  const correction = {
    id: 'a-corriger',
    motDePasse: 'demo1234',
    titre: 'Vélo de course révisé',
    description: 'Cadre carbone, photos ajoutées.',
    prix: 42000,
    categorie: 'loisirs',
    ville: 'Lyon',
    codePostal: '69003',
  }
  const rejected = async (promise: Promise<unknown>) =>
    promise.then(
      () => undefined,
      (error: unknown) => error,
    )

  it('gives a challengee annonce back with its motif for the mot de passe', async () => {
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
    })

    const edition = await gateway.editerAnnonce({
      id: 'a-corriger',
      motDePasse: 'demo1234',
    })

    expect(edition).toEqual({
      titre: 'Vélo de course',
      description: 'Cadre carbone, très bon état.',
      prix: 45000,
      categorie: 'loisirs',
      ville: 'Lyon',
      codePostal: '69003',
      motif: 'Photos manquantes',
    })
  })

  it('refuses a wrong mot de passe, naming the field', async () => {
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
    })

    const error = await rejected(
      gateway.editerAnnonce({ id: 'a-corriger', motDePasse: 'mauvais-mot' }),
    )

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 400, champ: 'motDePasse' })
  })

  it('answers 404 for an annonce that is unknown or not challengee', async () => {
    const annonce = uneAnnonce()
    const gateway = createInMemoryAnnonceGateway({
      annonces: [annonce],
      annoncesChallengees: [challengee],
    })

    for (const id of ['inconnue', annonce.id]) {
      const error = await rejected(
        gateway.editerAnnonce({ id, motDePasse: 'demo1234' }),
      )
      expect(error).toMatchObject({ status: 404 })
      expect(
        await rejected(gateway.corrigerAnnonce({ ...correction, id })),
      ).toMatchObject({ status: 404 })
    }
  })

  it('sends a corrected annonce back to moderation: it is no longer challengee', async () => {
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
    })

    await gateway.corrigerAnnonce(correction)

    expect(await gateway.consulterAnnonce('a-corriger')).toBeNull()
    expect(await rejected(gateway.corrigerAnnonce(correction))).toMatchObject({
      status: 404,
    })
  })

  it('refuses a correction with a wrong mot de passe and keeps the annonce challengee', async () => {
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
    })

    const error = await rejected(
      gateway.corrigerAnnonce({ ...correction, motDePasse: 'mauvais-mot' }),
    )

    expect(error).toMatchObject({ status: 400, champ: 'motDePasse' })
    await expect(
      gateway.editerAnnonce({ id: 'a-corriger', motDePasse: 'demo1234' }),
    ).resolves.toMatchObject({ titre: 'Vélo de course' })
  })

  it('refuses a correction the way the API would, when told to', async () => {
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
      correctionRefusee: { erreur: 'Le titre est trop court', champ: 'titre' },
    })

    const error = await rejected(gateway.corrigerAnnonce(correction))

    expect(error).toMatchObject({
      status: 400,
      champ: 'titre',
      message: 'Le titre est trop court',
    })
  })
})
