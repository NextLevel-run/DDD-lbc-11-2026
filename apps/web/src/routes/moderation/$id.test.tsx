import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import { createInMemoryModerationGateway } from '#/features/moderation/in-memory-moderation-gateway'
import { renderRoute } from '#/test/renderRoute'
import { uneAnnonceEnFile } from '#/test/moderation'
import type { AnnonceEnFile } from '#/features/moderation/in-memory-moderation-gateway'

const renderDetail = (
  annonce: AnnonceEnFile = uneAnnonceEnFile(),
  id = annonce.id,
) =>
  renderRoute(`/moderation/${id}`, {
    gateways: {
      moderation: createInMemoryModerationGateway({ annonces: [annonce] }),
    },
  })

describe("Modération d'une annonce page", () => {
  beforeEach(() => {
    window.localStorage.setItem('moderateur', 'alice')
  })

  it('shows the content and the history of the annonce', async () => {
    await renderDetail()

    expect(
      await screen.findByRole('heading', { name: 'Vélo de ville' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Très bon état.')).toBeInTheDocument()
    expect(screen.getByText(/Véhicules/)).toBeInTheDocument()
    expect(screen.getByText(/Mise en file de modération/)).toBeInTheDocument()
    expect(screen.getByText('Libre')).toBeInTheDocument()
  })

  it('locks the annonce then releases it', async () => {
    const user = userEvent.setup()
    await renderDetail()

    await user.click(await screen.findByRole('button', { name: 'Verrouiller' }))

    expect(await screen.findByText('Verrouillée par alice')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Verrouiller' })).toBeNull()

    await user.click(screen.getByRole('button', { name: 'Relâcher' }))

    expect(await screen.findByText('Libre')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Verrouiller' }),
    ).toBeInTheDocument()
  })

  it('shows the 409 in French when another moderator holds the lock', async () => {
    const user = userEvent.setup()
    await renderDetail(
      uneAnnonceEnFile({
        verrouillePar: 'bob',
        verrouilleDepuis: new Date().toISOString(),
      }),
    )

    expect(await screen.findByText('Verrouillée par bob')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Relâcher' })).toBeNull()
    await user.click(screen.getByRole('button', { name: 'Verrouiller' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Cette annonce est déjà verrouillée par un autre modérateur',
    )
  })

  it('takes over a lock older than two hours', async () => {
    const user = userEvent.setup()
    await renderDetail(
      uneAnnonceEnFile({
        verrouillePar: 'bob',
        verrouilleDepuis: new Date(
          Date.now() - 3 * 60 * 60 * 1000,
        ).toISOString(),
      }),
    )

    expect(await screen.findByText('Libre')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Verrouiller' }))

    expect(await screen.findByText('Verrouillée par alice')).toBeInTheDocument()
    expect(screen.getByText(/Verrou expiré après 2 heures/)).toBeInTheDocument()
  })

  describe('with the lock of the moderator', () => {
    const verrouillee = () =>
      uneAnnonceEnFile({
        verrouillePar: 'alice',
        verrouilleDepuis: new Date().toISOString(),
      })

    it('offers the decisions only to the holder of the lock', async () => {
      const user = userEvent.setup()
      await renderDetail()

      await user.click(
        await screen.findByRole('button', { name: 'Verrouiller' }),
      )

      expect(
        await screen.findByRole('button', { name: 'Valider' }),
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'Refuser' }),
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'Challenger' }),
      ).toBeInTheDocument()
    })

    it('offers no decision when the annonce is free or locked by another', async () => {
      await renderDetail(
        uneAnnonceEnFile({
          verrouillePar: 'bob',
          verrouilleDepuis: new Date().toISOString(),
        }),
      )

      await screen.findByText('Verrouillée par bob')
      expect(screen.queryByRole('button', { name: 'Valider' })).toBeNull()
      expect(screen.queryByRole('button', { name: 'Refuser' })).toBeNull()
      expect(screen.queryByRole('button', { name: 'Challenger' })).toBeNull()
    })

    it('validates the annonce and shows the decision', async () => {
      const user = userEvent.setup()
      await renderDetail(verrouillee())

      await user.click(await screen.findByRole('button', { name: 'Valider' }))

      expect(
        await screen.findByText(/annonce validée, elle est en ligne/),
      ).toBeInTheDocument()
      expect(screen.getByText(/Validée par alice/)).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: 'Valider' })).toBeNull()
      expect(screen.queryByRole('button', { name: 'Verrouiller' })).toBeNull()
    })

    it('refuses the annonce and shows the decision', async () => {
      const user = userEvent.setup()
      await renderDetail(verrouillee())

      await user.click(await screen.findByRole('button', { name: 'Refuser' }))

      expect(await screen.findByText(/annonce refusée/)).toBeInTheDocument()
      expect(screen.getByText(/Refusée par alice/)).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: 'Refuser' })).toBeNull()
    })

    it('challenges the annonce with a motif', async () => {
      const user = userEvent.setup()
      await renderDetail(verrouillee())

      await user.type(
        await screen.findByLabelText(/^Motif du challenge/),
        'Photos manquantes',
      )
      await user.click(screen.getByRole('button', { name: 'Challenger' }))

      expect(await screen.findByText(/annonce challengée/)).toBeInTheDocument()
      expect(
        screen.getByText(/Challengée par alice : Photos manquantes/),
      ).toBeInTheDocument()
      expect(screen.queryByLabelText(/^Motif du challenge/)).toBeNull()
    })

    it('requires a motif of 10 characters to challenge', async () => {
      const user = userEvent.setup()
      await renderDetail(verrouillee())

      await user.type(
        await screen.findByLabelText(/^Motif du challenge/),
        'court',
      )
      await user.click(screen.getByRole('button', { name: 'Challenger' }))

      expect(
        await screen.findByText(
          'Le motif doit comporter de 10 à 1 000 caractères',
        ),
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'Valider' }),
      ).toBeInTheDocument()
    })

    it('shows the 409 in French when the lock expired meanwhile', async () => {
      const user = userEvent.setup()
      let maintenant = Date.now()
      await renderRoute('/moderation/a1', {
        gateways: {
          moderation: createInMemoryModerationGateway({
            annonces: [verrouillee()],
            now: () => new Date(maintenant),
          }),
        },
      })
      const valider = await screen.findByRole('button', { name: 'Valider' })

      // The page loaded a valid lock; the decision comes three hours later.
      maintenant += 3 * 60 * 60 * 1000
      await user.click(valider)

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'Votre verrou a expiré',
      )
      expect(
        screen.getByRole('button', { name: 'Valider' }),
      ).toBeInTheDocument()
    })
  })

  it('says the annonce is introuvable', async () => {
    await renderDetail(uneAnnonceEnFile(), 'inconnue')

    expect(
      await screen.findByRole('heading', { name: 'Annonce introuvable' }),
    ).toBeInTheDocument()
  })

  it('asks for the pseudo when there is none yet', async () => {
    window.localStorage.clear()
    await renderDetail()

    expect(
      await screen.findByLabelText(/^Votre pseudo de modérateur/),
    ).toBeInTheDocument()
  })
})
