import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Link, createFileRoute } from '@tanstack/react-router'
import { Button } from '@spark-ui/components/button'
import { categories } from '#/features/annonce/categories'
import {
  corrigerAnnonceMutationOptions,
  editerAnnonceMutationOptions,
} from '#/features/annonce/queries'
import {
  centimesEnEuros,
  corrigerAnnonceSchema,
  motDePasseSchema,
} from '#/features/annonce/schemas'
import { useAppForm } from '#/hooks/form'
import { ApiError } from '#/lib/http'
import type { AnnonceAEditer } from '#/features/annonce/annonce-gateway'
import type { CorrigerAnnonceValues } from '#/features/annonce/schemas'

export const Route = createFileRoute('/annonces/$id_/corriger')({
  head: () => ({ meta: [{ title: 'Corriger mon annonce' }] }),
  component: CorrigerAnnoncePage,
})

type Etape =
  | { nom: 'motDePasse' }
  | { nom: 'correction'; motDePasse: string; edition: AnnonceAEditer }
  | { nom: 'corrigee' }
  | { nom: 'introuvable' }

// An ApiError that names one of the fields is shown under that field; any
// other error is shown above the button.
const erreurDeChamp = <TChamp extends string>(
  error: unknown,
  champs: ReadonlyArray<TChamp>,
) =>
  error instanceof ApiError &&
  error.champ &&
  champs.includes(error.champ as TChamp)
    ? { champ: error.champ as TChamp, message: error.message }
    : undefined

const erreurGenerale = (error: Error | null, champs: ReadonlyArray<string>) =>
  error && !erreurDeChamp(error, champs) ? error.message : undefined

function CorrigerAnnoncePage() {
  const [etape, setEtape] = useState<Etape>({ nom: 'motDePasse' })

  switch (etape.nom) {
    case 'motDePasse':
      return (
        <MotDePasseForm
          onEdition={(motDePasse, edition) =>
            setEtape({ nom: 'correction', motDePasse, edition })
          }
          onIntrouvable={() => setEtape({ nom: 'introuvable' })}
        />
      )
    case 'correction':
      return (
        <CorrectionForm
          motDePasse={etape.motDePasse}
          edition={etape.edition}
          onCorrigee={() => setEtape({ nom: 'corrigee' })}
          onIntrouvable={() => setEtape({ nom: 'introuvable' })}
        />
      )
    case 'corrigee':
      return <Confirmation />
    case 'introuvable':
      return <AnnonceIntrouvable />
  }
}

const champsMotDePasse = ['motDePasse'] as const

// The mot de passe comes first: it is what lets the vendeur see the annonce
// and its motif, and the API gives nothing away before it.
function MotDePasseForm({
  onEdition,
  onIntrouvable,
}: {
  onEdition: (motDePasse: string, edition: AnnonceAEditer) => void
  onIntrouvable: () => void
}) {
  const { gateways } = Route.useRouteContext()
  const { id } = Route.useParams()
  const edition = useMutation(editerAnnonceMutationOptions(gateways.annonce))

  const form = useAppForm({
    defaultValues: { motDePasse: '' },
    validators: { onChange: motDePasseSchema },
    listeners: {
      // An error of the API holds until the vendeur changes the field.
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
        const aEditer = await edition.mutateAsync({ id, ...value })
        onEdition(value.motDePasse, aEditer)
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          onIntrouvable()
          return
        }
        const erreur = erreurDeChamp(error, champsMotDePasse)
        if (erreur) {
          formApi.setFieldMeta(erreur.champ, (meta) => ({
            ...meta,
            errorMap: { ...meta.errorMap, onServer: erreur.message },
          }))
        }
      }
    },
  })

  return (
    <main className="page-wrap px-lg pt-2xl pb-2xl">
      <section className="mx-auto flex max-w-sz-640 flex-col gap-xl">
        <h1 className="text-display-2">Corriger mon annonce</h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Votre annonce a été renvoyée par la modération. Saisissez son mot de
          passe pour voir ce qui est demandé et la corriger.
        </p>
        <form
          noValidate
          className="flex flex-col gap-lg"
          onSubmit={(event) => {
            event.preventDefault()
            event.stopPropagation()
            void form.handleSubmit()
          }}
        >
          <form.AppField name="motDePasse">
            {(field) => (
              <field.PasswordField
                label="Mot de passe de l'annonce"
                description="Celui que vous avez choisi en déposant l'annonce."
                autoComplete="current-password"
                isRequired
              />
            )}
          </form.AppField>

          {erreurGenerale(edition.error, champsMotDePasse) && (
            <div
              role="alert"
              className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
            >
              {edition.error?.message}
            </div>
          )}

          <form.AppForm>
            <form.SubscribeButton label="Continuer" />
          </form.AppForm>
        </form>
      </section>
    </main>
  )
}

const champsCorrection = [
  'titre',
  'description',
  'prix',
  'categorie',
  'ville',
  'codePostal',
] as const

function CorrectionForm({
  motDePasse,
  edition,
  onCorrigee,
  onIntrouvable,
}: {
  motDePasse: string
  edition: AnnonceAEditer
  onCorrigee: () => void
  onIntrouvable: () => void
}) {
  const { gateways, queryClient } = Route.useRouteContext()
  const { id } = Route.useParams()
  const correction = useMutation(
    corrigerAnnonceMutationOptions(gateways.annonce, queryClient),
  )

  const valeursInitiales: CorrigerAnnonceValues = {
    titre: edition.titre,
    description: edition.description,
    prix: centimesEnEuros(edition.prix),
    categorie: edition.categorie,
    ville: edition.ville,
    codePostal: edition.codePostal,
  }

  const form = useAppForm({
    defaultValues: valeursInitiales,
    validators: { onChange: corrigerAnnonceSchema },
    listeners: {
      // An error of the API holds until the vendeur changes the field.
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
        await correction.mutateAsync({
          id,
          motDePasse,
          ...corrigerAnnonceSchema.parse(value),
        })
        onCorrigee()
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          onIntrouvable()
          return
        }
        const erreur = erreurDeChamp(error, champsCorrection)
        if (erreur) {
          formApi.setFieldMeta(erreur.champ, (meta) => ({
            ...meta,
            errorMap: { ...meta.errorMap, onServer: erreur.message },
          }))
        }
      }
    },
  })

  return (
    <main className="page-wrap px-lg pt-2xl pb-2xl">
      <section className="mx-auto flex max-w-sz-640 flex-col gap-xl">
        <h1 className="text-display-2">Corriger mon annonce</h1>
        <div className="flex flex-col gap-sm rounded-md bg-surface-inverse/dim-5 p-md">
          <p className="text-headline-2">Ce que la modération demande</p>
          <p className="text-body-1">{edition.motif}</p>
        </div>
        <form
          noValidate
          className="flex flex-col gap-xl"
          onSubmit={(event) => {
            event.preventDefault()
            event.stopPropagation()
            void form.handleSubmit()
          }}
        >
          <fieldset className="flex flex-col gap-lg">
            <legend className="mb-md text-headline-2">Votre annonce</legend>
            <form.AppField name="titre">
              {(field) => <field.TextField label="Titre" isRequired />}
            </form.AppField>
            <form.AppField name="description">
              {(field) => (
                <field.TextArea label="Description" rows={6} isRequired />
              )}
            </form.AppField>
            <form.AppField name="prix">
              {(field) => (
                <field.TextField
                  label="Prix (€)"
                  description="0 pour un don"
                  inputMode="decimal"
                  isRequired
                />
              )}
            </form.AppField>
            <form.AppField name="categorie">
              {(field) => (
                <field.Select
                  label="Catégorie"
                  placeholder="Choisissez une catégorie"
                  values={categories.map(({ code, libelle }) => ({
                    value: code,
                    label: libelle,
                  }))}
                  isRequired
                />
              )}
            </form.AppField>
          </fieldset>

          <fieldset className="flex flex-col gap-lg">
            <legend className="mb-md text-headline-2">Localisation</legend>
            <form.AppField name="ville">
              {(field) => (
                <field.TextField
                  label="Ville"
                  autoComplete="address-level2"
                  isRequired
                />
              )}
            </form.AppField>
            <form.AppField name="codePostal">
              {(field) => (
                <field.TextField
                  label="Code postal"
                  inputMode="numeric"
                  autoComplete="postal-code"
                  isRequired
                />
              )}
            </form.AppField>
          </fieldset>

          {erreurGenerale(correction.error, champsCorrection) && (
            <div
              role="alert"
              className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
            >
              {correction.error?.message}
            </div>
          )}

          <form.AppForm>
            <form.SubscribeButton label="Envoyer la correction" />
          </form.AppForm>
        </form>
      </section>
    </main>
  )
}

function Confirmation() {
  return (
    <main className="page-wrap px-lg pt-3xl pb-2xl">
      <section className="flex flex-col items-start gap-lg rounded-xl border-sm border-outline bg-surface px-xl py-3xl text-on-surface">
        <h1 className="text-display-2">
          Votre annonce a été corrigée, elle repart en modération
        </h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Elle sera en ligne dès qu'elle aura été validée.
        </p>
        <Button asChild>
          <Link to="/">Retour à l'accueil</Link>
        </Button>
      </section>
    </main>
  )
}

function AnnonceIntrouvable() {
  return (
    <main className="page-wrap px-lg pt-3xl pb-2xl">
      <section className="flex flex-col items-start gap-lg rounded-xl border-sm border-outline bg-surface px-xl py-3xl text-on-surface">
        <h1 className="text-display-2">Annonce introuvable</h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Cette annonce n'existe pas ou n'a pas à être corrigée.
        </p>
        <Button asChild>
          <Link to="/">Retour à l'accueil</Link>
        </Button>
      </section>
    </main>
  )
}
