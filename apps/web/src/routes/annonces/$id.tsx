import { useState } from 'react'
import { useMutation, useSuspenseQuery } from '@tanstack/react-query'
import { Link, createFileRoute, notFound } from '@tanstack/react-router'
import { Button } from '@spark-ui/components/button'
import { Card } from '@spark-ui/components/card'
import { Divider } from '@spark-ui/components/divider'
import { Tag } from '@spark-ui/components/tag'
import placeholder from '#/assets/annonce-placeholder.svg'
import { formatDeposeeLe, formatPrix } from '#/features/annonce/format'
import {
  annonceQueryOptions,
  contacterMutationOptions,
} from '#/features/annonce/queries'
import { contacterSchema } from '#/features/annonce/schemas'
import { useAppForm } from '#/hooks/form'
import { ApiError } from '#/lib/http'
import type { Annonce } from '#/features/annonce/annonce-gateway'
import type { ContacterValues } from '#/features/annonce/schemas'

export const Route = createFileRoute('/annonces/$id')({
  loader: async ({ context, params }) => {
    const annonce = await context.queryClient.ensureQueryData(
      annonceQueryOptions(context.gateways.annonce, params.id),
    )
    if (!annonce) {
      throw notFound()
    }
    return { titre: annonce.titre }
  },
  head: ({ loaderData }) => ({
    meta: [{ title: loaderData ? loaderData.titre : 'Annonce introuvable' }],
  }),
  component: AnnoncePage,
  notFoundComponent: AnnonceIntrouvable,
})

function AnnoncePage() {
  const { gateways } = Route.useRouteContext()
  const { id } = Route.useParams()
  const { data: annonce } = useSuspenseQuery(
    annonceQueryOptions(gateways.annonce, id),
  )

  if (!annonce) {
    return <AnnonceIntrouvable />
  }
  return <AnnonceDetail annonce={annonce} />
}

function AnnonceDetail({ annonce }: { annonce: Annonce }) {
  const { localisation } = annonce
  const image = annonce.images.at(0)

  return (
    <main className="page-wrap px-lg pt-2xl pb-2xl">
      <article className="grid gap-xl md:grid-cols-[3fr_2fr]">
        <div className="overflow-hidden rounded-xl bg-neutral-container">
          {image ? (
            <img
              src={image}
              alt={annonce.titre}
              className="aspect-[4/3] w-full object-cover"
            />
          ) : (
            <img
              src={placeholder}
              alt="Pas de photo"
              className="aspect-[4/3] w-full object-cover"
            />
          )}
        </div>

        <Card design="outlined" intent="surface">
          <Card.Content className="flex flex-col gap-lg">
            <Tag design="tinted" intent="support" className="self-start">
              {annonce.categorie.libelle}
            </Tag>
            <h1 className="text-display-3">{annonce.titre}</h1>
            <p className="text-headline-1 text-main">
              {formatPrix(annonce.prix)}
            </p>
            <p className="text-body-2 text-on-surface/dim-1">
              Déposée le {formatDeposeeLe(annonce.deposeeLe)}
            </p>

            <Divider />

            <dl className="grid grid-cols-[auto_1fr] gap-x-lg gap-y-md text-body-1">
              <dt className="text-on-surface/dim-1">Vendeur</dt>
              <dd className="text-body-1-highlight">
                {annonce.vendeur.pseudo}
              </dd>
              <dt className="text-on-surface/dim-1">Ville</dt>
              <dd>
                {localisation.ville} {localisation.codePostal}
              </dd>
              <dt className="text-on-surface/dim-1">Lieu</dt>
              <dd>
                {localisation.departement.nom} ({localisation.departement.code}
                ), {localisation.region.nom}
              </dd>
            </dl>
          </Card.Content>
        </Card>

        <section className="md:col-span-2">
          <h2 className="mb-md text-headline-2">Description</h2>
          <p className="whitespace-pre-line text-body-1">
            {annonce.description}
          </p>
        </section>

        <section className="md:col-span-2">
          <ContactVendeur id={annonce.id} />
        </section>
      </article>
    </main>
  )
}

// The button unfolds the contact form; once sent, a confirmation replaces it.
function ContactVendeur({ id }: { id: string }) {
  const [etape, setEtape] = useState<'ferme' | 'ouvert' | 'envoye'>('ferme')

  if (etape === 'envoye') {
    return (
      <p
        role="status"
        className="rounded-md bg-success-container p-md text-body-1 text-on-success-container"
      >
        Votre message a été envoyé au vendeur.
      </p>
    )
  }
  if (etape === 'ferme') {
    return (
      <Button onClick={() => setEtape('ouvert')}>Contacter le vendeur</Button>
    )
  }
  return <FormulaireContact id={id} onEnvoye={() => setEtape('envoye')} />
}

const valeursInitiales: ContacterValues = { pseudo: '', email: '', message: '' }

const champs = new Set<string>(Object.keys(valeursInitiales))

// An ApiError that names a field of the form is shown under that field;
// any other error (the email could not leave) is shown above the form.
const erreurDeChamp = (error: unknown) =>
  error instanceof ApiError && error.champ && champs.has(error.champ)
    ? { champ: error.champ as keyof ContacterValues, message: error.message }
    : undefined

function FormulaireContact({
  id,
  onEnvoye,
}: {
  id: string
  onEnvoye: () => void
}) {
  const { gateways } = Route.useRouteContext()
  const contact = useMutation(contacterMutationOptions(gateways.annonce))

  const form = useAppForm({
    defaultValues: valeursInitiales,
    validators: { onChange: contacterSchema },
    listeners: {
      // An error of the API holds until the acheteur changes the field.
      onChange: ({ fieldApi }) => {
        if (fieldApi.state.meta.errorMap.onServer !== undefined) {
          fieldApi.setMeta((meta) => ({
            ...meta,
            errorMap: { ...meta.errorMap, onServer: undefined },
          }))
        }
      },
    },
    onSubmit: async ({ value, formApi }) => {
      try {
        await contact.mutateAsync({ id, ...contacterSchema.parse(value) })
        onEnvoye()
      } catch (error) {
        const erreur = erreurDeChamp(error)
        if (erreur) {
          formApi.setFieldMeta(erreur.champ, (meta) => ({
            ...meta,
            errorMap: { ...meta.errorMap, onServer: erreur.message },
          }))
        }
      }
    },
  })

  const erreurGenerale =
    contact.error && !erreurDeChamp(contact.error)
      ? contact.error.message
      : undefined

  return (
    <div className="flex max-w-sz-640 flex-col gap-lg">
      <h2 className="text-headline-2">Contacter le vendeur</h2>
      {erreurGenerale && (
        <div
          role="alert"
          className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
        >
          {erreurGenerale}
        </div>
      )}
      <form
        noValidate
        className="flex flex-col gap-lg"
        onSubmit={(event) => {
          event.preventDefault()
          event.stopPropagation()
          void form.handleSubmit()
        }}
      >
        <form.AppField name="pseudo">
          {(field) => (
            <field.TextField
              label="Votre pseudo"
              autoComplete="nickname"
              isRequired
            />
          )}
        </form.AppField>
        <form.AppField name="email">
          {(field) => (
            <field.TextField
              label="Votre email"
              description="Le vendeur vous répondra à cette adresse."
              type="email"
              inputMode="email"
              autoComplete="email"
              isRequired
            />
          )}
        </form.AppField>
        <form.AppField name="message">
          {(field) => (
            <field.TextArea label="Votre message" rows={5} isRequired />
          )}
        </form.AppField>

        <form.AppForm>
          <form.SubscribeButton label="Envoyer le message" />
        </form.AppForm>
      </form>
    </div>
  )
}

function AnnonceIntrouvable() {
  return (
    <main className="page-wrap px-lg pt-3xl pb-2xl">
      <section className="flex flex-col items-start gap-lg rounded-xl border-sm border-outline bg-surface px-xl py-3xl text-on-surface">
        <h1 className="text-display-2">Annonce introuvable</h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Cette annonce n'existe pas ou n'est plus en ligne.
        </p>
        <Button asChild>
          <Link to="/">Retour à l'accueil</Link>
        </Button>
      </section>
    </main>
  )
}
