import { expect } from '@playwright/test'
import type { APIRequestContext, Page } from '@playwright/test'
import { API_URL, MAILCATCHER_URL } from '../ports'

// Every test runs against one shared, in-memory API: it builds its own
// annonces around a unique token, then searches by that token, so it never
// sees the annonces of another test.
let compteur = 0
export function unJeton(): string {
  compteur += 1
  return `e2e${Date.now().toString(36)}${process.pid}x${compteur}`
}

export interface Depot {
  titre: string
  description: string
  prix: number // centimes
  categorie: string
  ville: string
  codePostal: string
  pseudo: string
  email: string
  motDePasse: string
}

export function unDepot(jeton: string, depot: Partial<Depot> = {}): Depot {
  return {
    titre: `Lampe ${jeton}`,
    description: 'Lampe de chevet en laiton, très bon état.',
    prix: 2500,
    categorie: 'maison-jardin',
    ville: 'Lyon',
    codePostal: '69003',
    pseudo: 'vendeur-e2e',
    email: `${jeton}@example.com`,
    motDePasse: 'secret12',
    ...depot,
  }
}

// ouvrir loads a page and waits for React to hydrate it. The server renders
// the forms, so acting earlier submits them natively or clicks a button that
// has no handler yet. React marks the body with its fiber once hydrated.
export async function ouvrir(page: Page, url: string) {
  await page.goto(url)
  await page.waitForFunction(() =>
    Object.keys(document.body).some((k) => k.startsWith('__reactFiber')),
  )
}

// deposer sets up data straight through the API: the form itself is covered
// once, by the parcours test. A new annonce awaits moderation and is online
// nowhere: use deposerEnLigne for one the test wants to find.
export async function deposer(
  request: APIRequestContext,
  depot: Depot,
): Promise<string> {
  const response = await request.post(`${API_URL}/annonces`, { data: depot })
  expect(response.status(), await response.text()).toBe(201)
  const { id } = (await response.json()) as { id: string }
  return id
}

// deposerEnLigne deposits an annonce then has a moderator validate it through
// the API, and waits for the async consumers to put it online.
export async function deposerEnLigne(
  request: APIRequestContext,
  depot: Depot,
): Promise<string> {
  const id = await deposer(request, depot)
  const moderation = `${API_URL}/moderation/annonces/${id}`
  // The queue is fed by an async consumer.
  await expect
    .poll(async () => (await request.get(moderation)).status(), {
      message: `annonce ${id} in the moderation queue`,
    })
    .toBe(200)
  const moderateur = { data: { moderateur: 'e2e-moderateur' } }
  for (const action of ['verrou', 'validation']) {
    const response = await request.post(`${moderation}/${action}`, moderateur)
    expect(response.status(), await response.text()).toBe(204)
  }
  await expect
    .poll(
      async () => (await request.get(`${API_URL}/annonces/${id}`)).status(),
      {
        message: `annonce ${id} online`,
      },
    )
    .toBe(200)
  return id
}

export async function remplirLeDepot(page: Page, depot: Depot) {
  await page.getByLabel(/^Titre/).fill(depot.titre)
  await page.getByLabel(/^Description/).fill(depot.description)
  await page
    .getByLabel(/^Prix/)
    .fill((depot.prix / 100).toFixed(2).replace('.', ','))
  await page.getByLabel(/^Catégorie/).selectOption(depot.categorie)
  await page.getByLabel(/^Ville/).fill(depot.ville)
  await page.getByLabel(/^Code postal/).fill(depot.codePostal)
  await page.getByLabel(/^Pseudo/).fill(depot.pseudo)
  await page.getByLabel(/^Email/).fill(depot.email)
  await page.getByLabel(/^Mot de passe/).fill(depot.motDePasse)
}

interface Message {
  id: number
  recipients: Array<string>
  subject: string
}

export interface Email {
  subject: string
  body: string
}

// emailPour waits for the email Mailcatcher received for this recipient: the
// API sends it from an async consumer, after answering the dépôt. A recipient
// gets several emails along the moderation journey: `sujet` picks the one
// whose subject contains it.
export async function emailPour(
  request: APIRequestContext,
  destinataire: string,
  sujet = '',
): Promise<Email> {
  let message: Message | undefined
  await expect
    .poll(
      async () => {
        const response = await request.get(`${MAILCATCHER_URL}/messages`)
        const messages = (await response.json()) as Array<Message>
        message = messages.find(
          (m) =>
            m.recipients.some((r) => r.includes(destinataire)) &&
            m.subject.includes(sujet),
        )
        return message !== undefined
      },
      {
        message: `an email for ${destinataire} (${sujet || 'any subject'}) in Mailcatcher`,
      },
    )
    .toBe(true)
  const plain = await request.get(
    `${MAILCATCHER_URL}/messages/${message!.id}.plain`,
  )
  return { subject: message!.subject, body: await plain.text() }
}
