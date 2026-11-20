import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import { createInMemoryModerationGateway } from '#/features/moderation/in-memory-moderation-gateway'
import { renderRoute } from '#/test/renderRoute'
import { uneAnnonceEnFile } from '#/test/moderation'

const renderFile = (annonces = [uneAnnonceEnFile()]) =>
  renderRoute('/moderation', {
    gateways: { moderation: createInMemoryModerationGateway({ annonces }) },
  })

describe('Modération page', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })

  it('asks for the pseudo before showing the file, and keeps it in the browser', async () => {
    const user = userEvent.setup()
    await renderFile()

    expect(
      await screen.findByLabelText(/^Votre pseudo de modérateur/),
    ).toBeInTheDocument()
    expect(screen.queryByText('Vélo de ville')).toBeNull()

    await user.type(screen.getByLabelText(/^Votre pseudo/), 'alice')
    await user.click(screen.getByRole('button', { name: 'Continuer' }))

    expect(await screen.findByText('Vélo de ville')).toBeInTheDocument()
    expect(window.localStorage.getItem('moderateur')).toBe('alice')
  })

  it('refuses a pseudo of less than 2 characters', async () => {
    const user = userEvent.setup()
    await renderFile()

    await user.type(await screen.findByLabelText(/^Votre pseudo/), 'a')
    await user.click(screen.getByRole('button', { name: 'Continuer' }))

    expect(
      await screen.findByText('Le pseudo doit comporter de 2 à 30 caractères'),
    ).toBeInTheDocument()
    expect(window.localStorage.getItem('moderateur')).toBeNull()
  })

  it('lists the annonces oldest first and says who holds a lock', async () => {
    window.localStorage.setItem('moderateur', 'alice')
    await renderFile([
      uneAnnonceEnFile({
        id: 'recente',
        titre: 'Table de jardin',
        misEnFileLe: '2026-11-19T09:30:00Z',
      }),
      uneAnnonceEnFile({
        id: 'ancienne',
        titre: 'Vélo de ville',
        misEnFileLe: '2026-11-19T08:30:00Z',
        verrouillePar: 'bob',
        verrouilleDepuis: new Date().toISOString(),
      }),
    ])

    const liens = await screen.findAllByRole('link', {
      name: /Vélo de ville|Table de jardin/,
    })
    expect(liens.map((lien) => lien.textContent)).toEqual([
      'Vélo de ville',
      'Table de jardin',
    ])
    expect(screen.getByText('Verrouillée par bob')).toBeInTheDocument()
  })

  it('says the file is empty', async () => {
    window.localStorage.setItem('moderateur', 'alice')
    await renderFile([])

    expect(
      await screen.findByText('Aucune annonce à modérer.'),
    ).toBeInTheDocument()
  })

  it('has no link in the navigation and is noindex', async () => {
    window.localStorage.setItem('moderateur', 'alice')
    await renderFile()
    await screen.findByText('Vélo de ville')

    expect(
      screen.getByRole('banner').querySelector('a[href^="/moderation"]'),
    ).toBeNull()
    expect(
      document.head
        .querySelector('meta[name="robots"]')
        ?.getAttribute('content'),
    ).toContain('noindex')
  })
})
