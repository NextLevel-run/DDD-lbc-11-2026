import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { createInMemoryAnnonceGateway } from '#/features/annonce/in-memory-annonce-gateway'
import { renderRoute } from '#/test/renderRoute'
import type { Annonce } from '#/features/annonce/annonce-gateway'

const velo: Annonce = {
  id: '0b6f4a8e-3c1d-4f2a-9e5b-7d8c9a0b1c2d',
  titre: 'Vélo électrique de ville',
  description: 'Autonomie 60 km, révisé il y a un mois.',
  prix: 89000,
  categorie: { code: 'loisirs', libelle: 'Loisirs' },
  localisation: {
    ville: 'Saint-Étienne',
    codePostal: '42000',
    departement: { code: '42', nom: 'Loire' },
    region: { code: '84', nom: 'Auvergne-Rhône-Alpes' },
  },
  vendeur: { pseudo: 'stephanois' },
  images: [],
  deposeeLe: '2026-11-02T10:00:00Z',
}

const renderAnnonce = (id: string, annonces: Array<Annonce>) =>
  renderRoute(`/annonces/${id}`, {
    gateways: { annonce: createInMemoryAnnonceGateway({ annonces }) },
  })

// Intl separates thousands and the currency sign with narrow no-break spaces.
const prixText = (expected: string) => (text: string) =>
  text.replace(/\s/g, ' ') === expected

describe('Annonce page', () => {
  it('shows the detail of the annonce', async () => {
    await renderAnnonce(velo.id, [velo])

    expect(
      await screen.findByRole('heading', { name: 'Vélo électrique de ville' }),
    ).toBeInTheDocument()
    expect(screen.getByText(prixText('890 €'))).toBeInTheDocument()
    expect(screen.getByText('Loisirs')).toBeInTheDocument()
    expect(
      screen.getByText('Autonomie 60 km, révisé il y a un mois.'),
    ).toBeInTheDocument()
    expect(screen.getByText('Saint-Étienne 42000')).toBeInTheDocument()
    expect(
      screen.getByText('Loire (42), Auvergne-Rhône-Alpes'),
    ).toBeInTheDocument()
    expect(screen.getByText('stephanois')).toBeInTheDocument()
    expect(screen.getByText('Déposée le 2 novembre 2026')).toBeInTheDocument()
  })

  it('shows the local placeholder when the annonce has no image', async () => {
    await renderAnnonce(velo.id, [velo])

    expect(
      await screen.findByRole('img', { name: 'Pas de photo' }),
    ).toBeInTheDocument()
  })

  it('shows the first image when the annonce has some', async () => {
    const avecPhoto = { ...velo, images: ['https://example.com/velo.jpg'] }

    await renderAnnonce(velo.id, [avecPhoto])

    expect(
      await screen.findByRole('img', { name: 'Vélo électrique de ville' }),
    ).toHaveAttribute('src', 'https://example.com/velo.jpg')
  })

  it('says Don when the prix is zero', async () => {
    await renderAnnonce(velo.id, [{ ...velo, prix: 0 }])

    expect(await screen.findByText('Don')).toBeInTheDocument()
  })

  it('shows Annonce introuvable for an unknown annonce', async () => {
    await renderAnnonce('8e2d1c4b-0000-4000-8000-000000000000', [velo])

    expect(
      await screen.findByRole('heading', { name: 'Annonce introuvable' }),
    ).toBeInTheDocument()
  })

  describe('contact', () => {
    const renderContact = (options: { envoiImpossible?: boolean } = {}) =>
      renderRoute(`/annonces/${velo.id}`, {
        gateways: {
          annonce: createInMemoryAnnonceGateway({
            annonces: [velo],
            ...options,
          }),
        },
      })

    async function ouvrirEtRemplir(
      user: ReturnType<typeof userEvent.setup>,
      message = 'Est-il toujours disponible ?',
    ) {
      await user.click(
        await screen.findByRole('button', { name: 'Contacter le vendeur' }),
      )
      await user.type(screen.getByLabelText(/^Votre pseudo/), 'acheteur42')
      await user.type(
        screen.getByLabelText(/^Votre email/),
        'acheteur@example.com',
      )
      await user.type(screen.getByLabelText(/^Votre message/), message)
    }

    const envoyer = () =>
      screen.getByRole('button', { name: 'Envoyer le message' })

    it('unfolds the contact form when the acheteur clicks the button', async () => {
      const user = userEvent.setup()
      await renderContact()

      expect(screen.queryByLabelText(/^Votre message/)).toBeNull()
      await user.click(
        await screen.findByRole('button', { name: 'Contacter le vendeur' }),
      )

      expect(screen.getByLabelText(/^Votre pseudo/)).toBeInTheDocument()
      expect(screen.getByLabelText(/^Votre email/)).toBeInTheDocument()
      expect(screen.getByLabelText(/^Votre message/)).toBeInTheDocument()
    })

    it('replaces the form with a confirmation once sent', async () => {
      const user = userEvent.setup()
      await renderContact()

      await ouvrirEtRemplir(user)
      await user.click(envoyer())

      expect(
        await screen.findByText('Votre message a été envoyé au vendeur.'),
      ).toBeInTheDocument()
      expect(screen.queryByLabelText(/^Votre message/)).toBeNull()
      expect(
        screen.queryByRole('button', { name: 'Contacter le vendeur' }),
      ).toBeNull()
    })

    it('shows a too short message under its field and sends nothing', async () => {
      const user = userEvent.setup()
      await renderContact()

      await ouvrirEtRemplir(user, 'court')
      await user.click(envoyer())

      expect(
        await screen.findByText(
          'Le message doit comporter de 10 à 2 000 caractères',
        ),
      ).toBeInTheDocument()
      expect(
        screen.queryByText('Votre message a été envoyé au vendeur.'),
      ).toBeNull()
    })

    it('shows the send failure above the form and keeps what was typed', async () => {
      const user = userEvent.setup()
      await renderContact({ envoiImpossible: true })

      await ouvrirEtRemplir(user)
      await user.click(envoyer())

      expect(await screen.findByRole('alert')).toHaveTextContent(
        "Votre message n'a pas pu être envoyé, réessayez plus tard.",
      )
      expect(screen.getByLabelText(/^Votre message/)).toHaveValue(
        'Est-il toujours disponible ?',
      )
      expect(envoyer()).toBeInTheDocument()
    })
  })
})
