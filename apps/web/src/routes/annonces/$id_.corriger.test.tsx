import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { createInMemoryAnnonceGateway } from '#/features/annonce/in-memory-annonce-gateway'
import { renderRoute } from '#/test/renderRoute'
import type { AnnonceGateway } from '#/features/annonce/annonce-gateway'

const challengee = {
  id: 'a-corriger',
  titre: 'Vélo de course',
  description: 'Cadre carbone, très bon état.',
  prix: 45050,
  categorie: 'loisirs',
  ville: 'Lyon',
  codePostal: '69003',
  motif: 'Photos manquantes',
}

const renderCorrection = (
  annonce: AnnonceGateway = createInMemoryAnnonceGateway({
    annoncesChallengees: [challengee],
  }),
  id = challengee.id,
) => renderRoute(`/annonces/${id}/corriger`, { gateways: { annonce } })

async function saisirLeMotDePasse(
  user: ReturnType<typeof userEvent.setup>,
  motDePasse = 'demo1234',
) {
  await user.type(screen.getByLabelText(/^Mot de passe/), motDePasse)
  await user.click(screen.getByRole('button', { name: 'Continuer' }))
}

const envoyer = () =>
  screen.getByRole('button', { name: 'Envoyer la correction' })

describe('Corriger une annonce page', () => {
  it('asks for the mot de passe before showing anything of the annonce', async () => {
    await renderCorrection()

    expect(
      await screen.findByRole('heading', { name: 'Corriger mon annonce' }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/^Mot de passe/)).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Titre/)).toBeNull()
    expect(screen.queryByText('Photos manquantes')).toBeNull()
  })

  it('prefills the form with the annonce and shows the motif of the challenge', async () => {
    const user = userEvent.setup()
    await renderCorrection()
    await screen.findByLabelText(/^Mot de passe/)

    await saisirLeMotDePasse(user)

    expect(await screen.findByText('Photos manquantes')).toBeInTheDocument()
    expect(screen.getByLabelText(/^Titre/)).toHaveValue('Vélo de course')
    expect(screen.getByLabelText(/^Description/)).toHaveValue(
      'Cadre carbone, très bon état.',
    )
    expect(screen.getByLabelText(/^Prix/)).toHaveValue('450,50')
    expect(screen.getByLabelText(/^Catégorie/)).toHaveValue('loisirs')
    expect(screen.getByLabelText(/^Ville/)).toHaveValue('Lyon')
    expect(screen.getByLabelText(/^Code postal/)).toHaveValue('69003')
  })

  it('shows a wrong mot de passe under its field and lets the vendeur retry', async () => {
    const user = userEvent.setup()
    await renderCorrection()
    await screen.findByLabelText(/^Mot de passe/)

    await saisirLeMotDePasse(user, 'mauvais-mot')

    expect(
      await screen.findByText("Le mot de passe de l'annonce est incorrect"),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Titre/)).toBeNull()

    await user.clear(screen.getByLabelText(/^Mot de passe/))
    await saisirLeMotDePasse(user)

    expect(await screen.findByLabelText(/^Titre/)).toBeInTheDocument()
  })

  it('sends the correction and confirms the annonce goes back to moderation', async () => {
    const user = userEvent.setup()
    const gateway = createInMemoryAnnonceGateway({
      annoncesChallengees: [challengee],
    })
    await renderCorrection(gateway)
    await screen.findByLabelText(/^Mot de passe/)
    await saisirLeMotDePasse(user)
    await screen.findByLabelText(/^Titre/)

    await user.clear(screen.getByLabelText(/^Titre/))
    await user.type(screen.getByLabelText(/^Titre/), 'Vélo de course révisé')
    await user.click(envoyer())

    expect(
      await screen.findByText(
        'Votre annonce a été corrigée, elle repart en modération',
      ),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Titre/)).toBeNull()
    expect(
      screen.getByRole('link', { name: "Retour à l'accueil" }),
    ).toBeInTheDocument()
    await expect(
      gateway.editerAnnonce({ id: 'a-corriger', motDePasse: 'demo1234' }),
    ).rejects.toMatchObject({ status: 404 })
  })

  it('checks the fields as the dépôt does, before sending', async () => {
    const user = userEvent.setup()
    await renderCorrection()
    await screen.findByLabelText(/^Mot de passe/)
    await saisirLeMotDePasse(user)
    await screen.findByLabelText(/^Titre/)

    await user.clear(screen.getByLabelText(/^Titre/))
    await user.click(envoyer())

    expect(
      await screen.findByText('Le titre doit comporter de 3 à 100 caractères'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(
        'Votre annonce a été corrigée, elle repart en modération',
      ),
    ).toBeNull()
  })

  it('shows the refusal of the API under the field it names', async () => {
    const user = userEvent.setup()
    await renderCorrection(
      createInMemoryAnnonceGateway({
        annoncesChallengees: [challengee],
        correctionRefusee: {
          erreur: 'Le code postal ne désigne aucun département',
          champ: 'codePostal',
        },
      }),
    )
    await screen.findByLabelText(/^Mot de passe/)
    await saisirLeMotDePasse(user)
    await screen.findByLabelText(/^Titre/)

    await user.click(envoyer())

    expect(
      await screen.findByText('Le code postal ne désigne aucun département'),
    ).toBeInTheDocument()
  })

  it('says the annonce is not found when it is unknown or not challengee', async () => {
    const user = userEvent.setup()
    await renderCorrection(undefined, 'inconnue')
    await screen.findByLabelText(/^Mot de passe/)

    await saisirLeMotDePasse(user)

    expect(
      await screen.findByRole('heading', { name: 'Annonce introuvable' }),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Titre/)).toBeNull()
  })
})
