import { expect, test } from '@playwright/test'
import { deposerEnLigne, ouvrir, unDepot, unJeton } from './support'

// The search page against the real API: the HTTP gateway's query string, the
// API's normalisation and filters, the real table of régions. The rules
// themselves are unit-tested in the domain and on the in-memory gateway.

test('the filters typed in the page reach the API and stay in the URL', async ({
  page,
  request,
}) => {
  const jeton = unJeton()
  const cherche = `Planche de surf ${jeton}`
  await Promise.all([
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: cherche,
        prix: 22000,
        categorie: 'loisirs',
        ville: 'Saint-Denis',
        codePostal: '97400',
      }),
    ),
    // The homonym in Seine-Saint-Denis.
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: `Planche à repasser ${jeton}`,
        prix: 22000,
        categorie: 'loisirs',
        ville: 'Saint-Denis',
        codePostal: '93200',
      }),
    ),
    // Out of the fourchette de prix.
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: `Planche de surf pro ${jeton}`,
        prix: 90000,
        categorie: 'loisirs',
        ville: 'Saint-Denis',
        codePostal: '97400',
      }),
    ),
    // Another catégorie.
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: `Planche de bois ${jeton}`,
        prix: 22000,
        categorie: 'maison-jardin',
        ville: 'Saint-Denis',
        codePostal: '97400',
      }),
    ),
  ])

  await ouvrir(page, `/recherche?motCle=${jeton}`)
  await expect(page.getByRole('heading', { name: '4 annonces' })).toBeVisible()

  await page.getByLabel('Catégorie').selectOption('loisirs')
  await page.getByLabel('Région ou département').selectOption('region:04')
  await page.getByLabel('Ville ou code postal').fill('saint denis')
  await page.getByLabel('Prix min (€)').fill('100')
  await page.getByLabel('Prix max (€)').fill('300')
  await page.getByRole('button', { name: 'Rechercher' }).click()

  await expect(page.getByRole('heading', { name: '1 annonce' })).toBeVisible()
  await expect(page.getByText(cherche)).toBeVisible()
  const url = new URL(page.url())
  expect(Object.fromEntries(url.searchParams)).toMatchObject({
    motCle: jeton,
    categorie: 'loisirs',
    region: '04',
    villeOuCodePostal: 'saint denis',
    prixMin: '100',
    prixMax: '300',
  })

  // The link is shareable: a fresh page on the same URL finds the same.
  const partage = await page.context().newPage()
  await ouvrir(partage, url.toString())
  await expect(
    partage.getByRole('heading', { name: '1 annonce' }),
  ).toBeVisible()
  await expect(partage.getByText(cherche)).toBeVisible()
})

test('the home search bar searches by catégorie and lieu', async ({
  page,
  request,
}) => {
  const jeton = unJeton()
  await Promise.all([
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: `Scooter ${jeton}`,
        categorie: 'vehicules',
        ville: 'Ajaccio',
        codePostal: '20000',
      }),
    ),
    deposerEnLigne(
      request,
      unDepot(jeton, {
        titre: `Scooter bastiais ${jeton}`,
        categorie: 'vehicules',
        ville: 'Bastia',
        codePostal: '20200',
      }),
    ),
  ])

  await ouvrir(page, '/')
  await page.getByLabel('Mot-clé').fill(jeton)
  await page.getByLabel('Catégorie').selectOption('vehicules')
  await page.getByLabel('Région ou département').selectOption('departement:2A')
  await page.getByRole('button', { name: 'Rechercher' }).click()

  await expect(page).toHaveURL(/\/recherche\?/)
  await expect(page.getByRole('heading', { name: '1 annonce' })).toBeVisible()
  await expect(page.getByText(`Scooter ${jeton}`)).toBeVisible()
})

test('the lieu selector offers every région and département of the API', async ({
  page,
}) => {
  await ouvrir(page, '/recherche')

  const lieu = page.getByLabel('Région ou département')
  // The in-memory gateway only knows a subset: these come from GET /regions.
  for (const valeur of [
    'region:84',
    'departement:2B',
    'departement:976',
    'region:01',
  ]) {
    await expect(lieu.locator(`option[value="${valeur}"]`)).toHaveCount(1)
  }
  await expect(lieu.locator('optgroup')).toHaveCount(18)
})

test('the pagination serves the next page from the API', async ({
  page,
  request,
}) => {
  const jeton = unJeton()
  for (let i = 1; i <= 21; i++) {
    // One at a time: the API dates the annonces, the first one is the oldest.
    await deposerEnLigne(
      request,
      unDepot(jeton, { titre: `Lot ${i} ${jeton}` }),
    )
  }

  await ouvrir(page, `/recherche?motCle=${jeton}`)
  await expect(page.getByRole('heading', { name: '21 annonces' })).toBeVisible()
  await expect(page.getByText(`Lot 21 ${jeton}`)).toBeVisible()
  await expect(page.getByText(`Lot 1 ${jeton}`, { exact: true })).toBeHidden()

  await page
    .getByRole('navigation', { name: 'Pagination' })
    .getByRole('button', { name: '2' })
    .click()

  await expect(page).toHaveURL(/page=2/)
  await expect(page.getByText(`Lot 1 ${jeton}`, { exact: true })).toBeVisible()
  await expect(page.getByText(`Lot 21 ${jeton}`)).toBeHidden()
})
