import { expect, test } from '@playwright/test'
import { API_URL } from '../ports'
import {
  deposer,
  emailPour,
  ouvrir,
  remplirLeDepot,
  unDepot,
  unJeton,
} from './support'

// The whole journey of a vendeur on the real stack: the form, the API, the
// async email through SMTP, the invisibility of the new annonce. Each page
// is unit-tested on in-memory gateways; this checks they hold together.
test('a vendeur deposits an annonce that awaits moderation and gets the email', async ({
  page,
  request,
}) => {
  const jeton = unJeton()
  const depot = unDepot(jeton, {
    titre: `Vélo électrique ${jeton}`,
    prix: 89050,
    categorie: 'loisirs',
    ville: 'Ajaccio',
    codePostal: '20090',
  })

  // Dépôt, from the header button.
  await ouvrir(page, '/')
  await page.getByRole('link', { name: 'Déposer une annonce' }).click()
  await remplirLeDepot(page, depot)
  await page.getByRole('button', { name: "Publier l'annonce" }).click()

  await expect(page.getByText('Annonce déposée')).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'Votre annonce est en cours de modération',
    }),
  ).toBeVisible()

  // Awaiting moderation, the annonce is online nowhere.
  await ouvrir(page, '/')
  await expect(page.getByText(depot.titre)).toHaveCount(0)
  await ouvrir(page, `/recherche?motCle=${jeton}`)
  await expect(page.getByRole('heading', { name: '0 annonce' })).toBeVisible()

  // The confirmation email, through Mailcatcher: it says the annonce awaits
  // moderation and keeps its links.
  const email = await emailPour(request, depot.email)
  expect(email.subject).toBe(
    `Votre annonce « ${depot.titre} » est en cours de modération`,
  )
  expect(email.body).not.toContain(depot.motDePasse)
  const urlAnnonce = email.body.match(/http\S+\/annonces\/[0-9a-f-]{36}/)?.[0]
  expect(urlAnnonce).toBeDefined()
  const lienSuppression = email.body.match(/http\S+\/supprimer/)?.[0]
  expect(lienSuppression).toBe(`${urlAnnonce}/supprimer`)

  // The links of the email lead nowhere yet: the annonce does not leak.
  await ouvrir(page, urlAnnonce!)
  await expect(
    page.getByRole('heading', { name: 'Annonce introuvable' }),
  ).toBeVisible()
  await ouvrir(page, lienSuppression!)
  await expect(
    page.getByRole('heading', { name: 'Annonce introuvable' }),
  ).toBeVisible()
})

// The whole correction journey on the real stack: a moderator challenges a
// deposited annonce on the hidden pages, the vendeur follows the link of the
// email, proves ownership, corrects, and the annonce goes back to the queue
// until a second validation puts it online. The dépôt form is covered above:
// the annonce is deposited through the API here.
test('a vendeur corrects a challenged annonce and it goes online once validated again', async ({
  page,
  request,
}) => {
  const jeton = unJeton()
  const depot = unDepot(jeton, { titre: `Table de jardin ${jeton}` })
  const id = await deposer(request, depot)
  const titreCorrige = `Table de jardin en teck ${jeton}`

  // The annonce reaches the queue through the async consumer.
  await expect
    .poll(async () =>
      (await request.get(`${API_URL}/moderation/annonces/${id}`)).status(),
    )
    .toBe(200)

  // Moderation: pseudo, lock, challenge with a motif.
  await ouvrir(page, `/moderation/${id}`)
  await page.getByLabel('Votre pseudo de modérateur').fill('alice')
  await page.getByRole('button', { name: 'Continuer' }).click()
  await page.getByRole('button', { name: 'Verrouiller' }).click()
  await page
    .getByLabel(/^Motif du challenge/)
    .fill('Ajoutez les dimensions et une photo de la table.')
  await page.getByRole('button', { name: 'Challenger' }).click()
  await expect(
    page.getByText(
      'Décision : annonce challengée, le vendeur doit la corriger.',
    ),
  ).toBeVisible()

  // The email carries the motif and the link to the correction page.
  const email = await emailPour(request, depot.email, 'doit être corrigée')
  expect(email.body).toContain('Ajoutez les dimensions et une photo')
  expect(email.body).not.toContain(depot.motDePasse)
  const lien = email.body.match(
    /http\S+\/annonces\/[0-9a-f-]{36}\/corriger/,
  )?.[0]
  expect(lien).toBeDefined()

  // The correction page: mot de passe first, then the prefilled form with the
  // motif.
  await ouvrir(page, lien!)
  await page.getByLabel(/^Mot de passe/).fill('mauvais-mot')
  await page.getByRole('button', { name: 'Continuer' }).click()
  await expect(
    page.getByText("Le mot de passe de l'annonce est incorrect"),
  ).toBeVisible()
  await page.getByLabel(/^Mot de passe/).fill(depot.motDePasse)
  await page.getByRole('button', { name: 'Continuer' }).click()
  await expect(
    page.getByText('Ajoutez les dimensions et une photo de la table.'),
  ).toBeVisible()
  await expect(page.getByLabel(/^Titre/)).toHaveValue(depot.titre)
  await expect(page.getByLabel(/^Prix/)).toHaveValue('25')
  await page.getByLabel(/^Titre/).fill(titreCorrige)
  await page.getByRole('button', { name: 'Envoyer la correction' }).click()
  await expect(
    page.getByText('Votre annonce a été corrigée, elle repart en modération'),
  ).toBeVisible()

  // Back in the queue with its whole history, and still offline.
  await expect
    .poll(async () => {
      const response = await request.get(`${API_URL}/moderation/annonces/${id}`)
      const detail = (await response.json()) as { statut: string }
      return detail.statut
    })
    .toBe('en-file')
  await ouvrir(page, `/annonces/${id}`)
  await expect(
    page.getByRole('heading', { name: 'Annonce introuvable' }),
  ).toBeVisible()
  await ouvrir(page, `/moderation/${id}`)
  await expect(page.getByRole('heading', { name: titreCorrige })).toBeVisible()
  await expect(page.getByText(/Challengée par alice : Ajoutez/)).toBeVisible()

  // Second moderation: validation puts the corrected annonce online.
  await page.getByRole('button', { name: 'Verrouiller' }).click()
  await page.getByRole('button', { name: 'Valider' }).click()
  await expect(
    page.getByText('Décision : annonce validée, elle est en ligne.'),
  ).toBeVisible()
  await expect
    .poll(async () => (await request.get(`${API_URL}/annonces/${id}`)).status())
    .toBe(200)
  await ouvrir(page, `/recherche?motCle=${jeton}`)
  await expect(page.getByText(titreCorrige)).toBeVisible()
  await ouvrir(page, `/annonces/${id}`)
  await expect(page.getByRole('heading', { name: titreCorrige })).toBeVisible()

  // The link of the email no longer opens a correction.
  await ouvrir(page, lien!)
  await page.getByLabel(/^Mot de passe/).fill(depot.motDePasse)
  await page.getByRole('button', { name: 'Continuer' }).click()
  await expect(
    page.getByRole('heading', { name: 'Annonce introuvable' }),
  ).toBeVisible()
})
