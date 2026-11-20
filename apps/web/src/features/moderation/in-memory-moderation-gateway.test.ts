import { describe, expect, it } from 'vitest'
import { ApiError } from '#/lib/http'
import { uneAnnonceEnFile } from '#/test/moderation'
import { createInMemoryModerationGateway } from './in-memory-moderation-gateway'
import type { AnnonceEnFile } from './in-memory-moderation-gateway'

const maintenant = new Date('2026-11-19T10:00:00Z')

const gateway = (annonces: Array<AnnonceEnFile>, now = maintenant) =>
  createInMemoryModerationGateway({ annonces, now: () => now })

describe('in-memory moderation gateway', () => {
  it('lists the file oldest first', async () => {
    const liste = await gateway([
      uneAnnonceEnFile({ id: 'recente', misEnFileLe: '2026-11-19T09:30:00Z' }),
      uneAnnonceEnFile({ id: 'ancienne', misEnFileLe: '2026-11-19T08:30:00Z' }),
    ]).listerAnnoncesAModerer()

    expect(liste.map(({ id }) => id)).toEqual(['ancienne', 'recente'])
  })

  it('locks an annonce for a moderator, shown in the list and the detail', async () => {
    const moderation = gateway([uneAnnonceEnFile()])

    await moderation.verrouillerAnnonce({ id: 'a1', moderateur: 'alice' })

    const [resume] = await moderation.listerAnnoncesAModerer()
    expect(resume.verrouillePar).toBe('alice')
    const detail = await moderation.consulterAnnonceAModerer('a1')
    expect(detail?.statut).toBe('verrouillee')
    expect(detail?.historique.map(({ type }) => type)).toEqual([
      'AnnonceMiseEnFile',
      'AnnonceVerrouillee',
    ])
  })

  it('refuses the lock of another moderator with a 409, unless it expired', async () => {
    const annonce = uneAnnonceEnFile({
      verrouillePar: 'alice',
      verrouilleDepuis: '2026-11-19T09:00:00Z',
    })

    await expect(
      gateway([annonce]).verrouillerAnnonce({ id: 'a1', moderateur: 'bob' }),
    ).rejects.toMatchObject({ status: 409 })

    const tard = new Date('2026-11-19T11:00:01Z')
    const moderation = gateway([annonce], tard)
    expect((await moderation.listerAnnoncesAModerer())[0].verrouillePar).toBe(
      '',
    )
    await moderation.verrouillerAnnonce({ id: 'a1', moderateur: 'bob' })
    const detail = await moderation.consulterAnnonceAModerer('a1')
    expect(detail?.verrouillePar).toBe('bob')
    expect(detail?.historique.map(({ type }) => type)).toContain(
      'VerrouAnnonceRelache',
    )
  })

  it('releases the lock of its holder only', async () => {
    const moderation = gateway([
      uneAnnonceEnFile({
        verrouillePar: 'alice',
        verrouilleDepuis: '2026-11-19T09:30:00Z',
      }),
    ])

    await expect(
      moderation.relacherVerrou({ id: 'a1', moderateur: 'bob' }),
    ).rejects.toMatchObject({ status: 409 })
    await moderation.relacherVerrou({ id: 'a1', moderateur: 'alice' })

    expect((await moderation.consulterAnnonceAModerer('a1'))?.statut).toBe(
      'en-file',
    )
  })

  it('answers 400 on moderateur and 404 for an unknown annonce', async () => {
    const moderation = gateway([uneAnnonceEnFile()])

    await expect(
      moderation.verrouillerAnnonce({ id: 'a1', moderateur: 'a' }),
    ).rejects.toMatchObject({ status: 400, champ: 'moderateur' })
    await expect(
      moderation.verrouillerAnnonce({ id: 'inconnue', moderateur: 'alice' }),
    ).rejects.toBeInstanceOf(ApiError)
    expect(await moderation.consulterAnnonceAModerer('inconnue')).toBeNull()
  })
})

describe('in-memory moderation gateway decisions', () => {
  const verrouille = () =>
    uneAnnonceEnFile({
      verrouillePar: 'alice',
      verrouilleDepuis: '2026-11-19T09:30:00Z',
    })

  it.each([
    ['validerAnnonce', 'validee', 'AnnonceValidee'],
    ['refuserAnnonce', 'refusee', 'AnnonceRefusee'],
  ] as const)(
    '%s takes the annonce out of the file and records the decision',
    async (decision, statut, type) => {
      const moderation = gateway([verrouille()])

      await moderation[decision]({ id: 'a1', moderateur: 'alice' })

      expect(await moderation.listerAnnoncesAModerer()).toEqual([])
      const detail = await moderation.consulterAnnonceAModerer('a1')
      expect(detail?.statut).toBe(statut)
      expect(detail?.verrouillePar).toBe('')
      expect(detail?.historique.at(-1)).toMatchObject({
        type,
        moderateur: 'alice',
      })
    },
  )

  it('challenges with a trimmed motif kept in the history', async () => {
    const moderation = gateway([verrouille()])

    await moderation.challengerAnnonce({
      id: 'a1',
      moderateur: 'alice',
      motif: '  Photos manquantes  ',
    })

    const detail = await moderation.consulterAnnonceAModerer('a1')
    expect(detail?.statut).toBe('challengee')
    expect(detail?.historique.at(-1)).toMatchObject({
      type: 'AnnonceChallengee',
      motif: 'Photos manquantes',
    })
  })

  it('refuses a decision without the valid lock of the moderator with a 409', async () => {
    const libre = gateway([uneAnnonceEnFile()])
    const decisions = (moderation: typeof libre, moderateur: string) => [
      moderation.validerAnnonce({ id: 'a1', moderateur }),
      moderation.refuserAnnonce({ id: 'a1', moderateur }),
      moderation.challengerAnnonce({
        id: 'a1',
        moderateur,
        motif: 'Photos manquantes',
      }),
    ]

    for (const decision of decisions(libre, 'alice')) {
      await expect(decision).rejects.toMatchObject({ status: 409 })
    }
    for (const decision of decisions(gateway([verrouille()]), 'bob')) {
      await expect(decision).rejects.toMatchObject({ status: 409 })
    }
    const tard = new Date('2026-11-19T11:30:01Z')
    for (const decision of decisions(gateway([verrouille()], tard), 'alice')) {
      await expect(decision).rejects.toMatchObject({
        status: 409,
        message: 'Votre verrou a expiré',
      })
    }
  })

  it('refuses an invalid motif with a 400 on motif, and decides nothing', async () => {
    const moderation = gateway([verrouille()])

    for (const motif of ['', 'court', '   neuf    ', 'x'.repeat(1001)]) {
      await expect(
        moderation.challengerAnnonce({ id: 'a1', moderateur: 'alice', motif }),
      ).rejects.toMatchObject({ status: 400, champ: 'motif' })
    }
    expect((await moderation.consulterAnnonceAModerer('a1'))?.statut).toBe(
      'verrouillee',
    )
  })

  it('does not decide twice nor lock a decided annonce', async () => {
    const moderation = gateway([verrouille()])
    await moderation.validerAnnonce({ id: 'a1', moderateur: 'alice' })

    await expect(
      moderation.refuserAnnonce({ id: 'a1', moderateur: 'alice' }),
    ).rejects.toMatchObject({ status: 409 })
    await expect(
      moderation.verrouillerAnnonce({ id: 'a1', moderateur: 'alice' }),
    ).rejects.toMatchObject({ status: 409 })
    await expect(
      moderation.validerAnnonce({ id: 'inconnue', moderateur: 'alice' }),
    ).rejects.toMatchObject({ status: 404 })
  })
})
