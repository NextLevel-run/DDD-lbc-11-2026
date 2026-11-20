import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { createInMemoryAnnonceGateway } from '#/features/annonce/in-memory-annonce-gateway'
import { renderRoute } from '#/test/renderRoute'
import type { AnnonceGateway } from '#/features/annonce/annonce-gateway'

const renderDepot = (
  annonce: AnnonceGateway = createInMemoryAnnonceGateway(),
) => renderRoute('/annonces/deposer', { gateways: { annonce } })

async function remplirLeFormulaire(
  user: ReturnType<typeof userEvent.setup>,
  prix = '12,50',
  motDePasse = 'demo1234',
) {
  await user.type(screen.getByLabelText(/^Titre/), 'Vélo de course')
  await user.type(
    screen.getByLabelText(/^Description/),
    'Cadre carbone, très bon état.',
  )
  await user.type(screen.getByLabelText(/^Prix/), prix)
  await user.selectOptions(screen.getByLabelText(/^Catégorie/), 'loisirs')
  await user.type(screen.getByLabelText(/^Ville/), 'Lyon')
  await user.type(screen.getByLabelText(/^Code postal/), '69003')
  await user.type(screen.getByLabelText(/^Pseudo/), 'cycliste69')
  await user.type(screen.getByLabelText(/^Email/), 'cycliste@example.com')
  await user.type(screen.getByLabelText(/^Mot de passe/), motDePasse)
}

const publier = () => screen.getByRole('button', { name: "Publier l'annonce" })

describe('Déposer une annonce page', () => {
  it('is reached from the header button', async () => {
    const user = userEvent.setup()
    await renderRoute('/')

    await user.click(screen.getByRole('link', { name: 'Déposer une annonce' }))

    expect(
      await screen.findByRole('heading', { name: 'Déposer une annonce' }),
    ).toBeInTheDocument()
  })

  it('deposits the annonce then says it is awaiting moderation', async () => {
    const user = userEvent.setup()
    await renderDepot()

    await remplirLeFormulaire(user)
    await user.click(publier())

    expect(
      await screen.findByRole('heading', {
        name: 'Votre annonce est en cours de modération',
      }),
    ).toBeInTheDocument()
    expect(await screen.findByText('Annonce déposée')).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: "Retour à l'accueil" }),
    ).toBeInTheDocument()
  })

  it('leaves the new annonce out of the consultation and the search', async () => {
    const user = userEvent.setup()
    const gateway = createInMemoryAnnonceGateway()
    const ids: Array<string> = []
    await renderDepot({
      ...gateway,
      deposerAnnonce: async (args) => {
        const id = await gateway.deposerAnnonce(args)
        ids.push(id)
        return id
      },
    })

    await remplirLeFormulaire(user)
    await user.click(publier())
    await screen.findByText('Votre annonce est en cours de modération')

    expect(await gateway.consulterAnnonce(ids[0])).toBeNull()
    expect((await gateway.rechercherAnnonces({})).total).toBe(0)
  })

  it('sends the prix in centimes and accepts a don', async () => {
    const user = userEvent.setup()
    const deposees: Array<number> = []
    const gateway = createInMemoryAnnonceGateway()
    await renderDepot({
      ...gateway,
      deposerAnnonce: (args) => {
        deposees.push(args.prix)
        return gateway.deposerAnnonce(args)
      },
    })

    await remplirLeFormulaire(user, '0')
    await user.click(publier())

    expect(
      await screen.findByText('Votre annonce est en cours de modération'),
    ).toBeInTheDocument()
    expect(deposees).toEqual([0])
  })

  it('shows the validation errors under their fields before sending', async () => {
    const user = userEvent.setup()
    await renderDepot()

    await user.type(screen.getByLabelText(/^Titre/), 'ab')
    await user.type(screen.getByLabelText(/^Telephone|^Téléphone/), '12345')
    await user.click(publier())

    expect(
      await screen.findByText('Le titre doit comporter de 3 à 100 caractères'),
    ).toBeInTheDocument()
    expect(
      screen.getByText(
        'La description doit comporter de 10 à 4 000 caractères',
      ),
    ).toBeInTheDocument()
    expect(screen.getByText('Le prix est obligatoire')).toBeInTheDocument()
    expect(screen.getByText('La catégorie est obligatoire')).toBeInTheDocument()
    expect(screen.getByText("L'email n'est pas valide")).toBeInTheDocument()
    expect(
      screen.getByText(
        'Le téléphone doit comporter 10 chiffres et commencer par 0',
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Déposer une annonce' }),
    ).toBeInTheDocument()
  })

  it('shows an error of the API under the field it names', async () => {
    const user = userEvent.setup()
    await renderDepot(
      createInMemoryAnnonceGateway({
        depotRefuse: {
          erreur: 'Le code postal ne désigne aucun département',
          champ: 'codePostal',
        },
      }),
    )

    await remplirLeFormulaire(user)
    await user.click(publier())

    const erreur = await screen.findByText(
      'Le code postal ne désigne aucun département',
    )
    expect(erreur).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Déposer une annonce' }),
    ).toBeInTheDocument()

    await user.type(screen.getByLabelText(/^Code postal/), '0')
    expect(
      screen.queryByText('Le code postal ne désigne aucun département'),
    ).not.toBeInTheDocument()
  })

  it('hides the mot de passe until the vendeur shows it', async () => {
    const user = userEvent.setup()
    await renderDepot()
    const champ = screen.getByLabelText(/^Mot de passe/)

    expect(champ).toHaveAttribute('type', 'password')
    await user.click(screen.getByRole('button', { name: 'Afficher' }))
    expect(champ).toHaveAttribute('type', 'text')
    await user.click(screen.getByRole('button', { name: 'Masquer' }))
    expect(champ).toHaveAttribute('type', 'password')
  })

  it('asks for a mot de passe of at least 8 characters under the field', async () => {
    const user = userEvent.setup()
    await renderDepot()

    await remplirLeFormulaire(user, '12,50', 'court')
    await user.click(publier())

    expect(
      await screen.findByText(
        "Le mot de passe de l'annonce doit comporter au moins 8 caractères",
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Déposer une annonce' }),
    ).toBeInTheDocument()
  })

  it('has no confirmation field for the mot de passe', async () => {
    await renderDepot()

    expect(screen.getAllByLabelText(/mot de passe/i)).toHaveLength(1)
  })

  it('shows an error of the API that names no field above the button', async () => {
    const user = userEvent.setup()
    await renderDepot(
      createInMemoryAnnonceGateway({
        depotRefuse: { erreur: 'Erreur interne' },
      }),
    )

    await remplirLeFormulaire(user)
    await user.click(publier())

    const alerte = await screen.findByRole('alert')
    expect(within(alerte).getByText('Erreur interne')).toBeInTheDocument()
  })
})
