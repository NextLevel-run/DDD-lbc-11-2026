import { expect, test } from '@playwright/test'
import { deposerEnLigne, unDepot, unJeton } from './support'

// Server-side rendering calls the API at API_URL, not through the Vite proxy:
// with JavaScript off, only that path can fill the page. Unit tests render in
// jsdom and never reach it.
test.use({ javaScriptEnabled: false })

test('an annonce is rendered on the server', async ({ page, request }) => {
  const jeton = unJeton()
  const depot = unDepot(jeton, { prix: 0 })
  const id = await deposerEnLigne(request, depot)

  await page.goto(`/annonces/${id}`)

  await expect(page.getByRole('heading', { name: depot.titre })).toBeVisible()
  await expect(page.getByText('Don', { exact: true })).toBeVisible()
  await expect(page.getByText('Rhône (69), Auvergne-Rhône-Alpes')).toBeVisible()
  // The vendeur shows by pseudo only.
  await expect(page.getByText(depot.pseudo)).toBeVisible()
  await expect(page.locator('body')).not.toContainText(depot.email)
})

test('an unknown annonce is rendered as introuvable', async ({ page }) => {
  await page.goto('/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f')

  await expect(
    page.getByRole('heading', { name: 'Annonce introuvable' }),
  ).toBeVisible()
})

test('search results are rendered on the server', async ({ page, request }) => {
  const jeton = unJeton()
  const depot = unDepot(jeton)
  await deposerEnLigne(request, depot)

  await page.goto(`/recherche?motCle=${jeton.toUpperCase()}`)

  await expect(page.getByRole('heading', { name: '1 annonce' })).toBeVisible()
  await expect(page.getByText(depot.titre)).toBeVisible()
})
