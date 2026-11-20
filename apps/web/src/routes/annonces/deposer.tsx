import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { Button } from '@spark-ui/components/button'
import { useToastManager } from '@spark-ui/components/toast'
import { categories } from '#/features/annonce/categories'
import { deposerAnnonceMutationOptions } from '#/features/annonce/queries'
import { deposerAnnonceSchema } from '#/features/annonce/schemas'
import { useAppForm } from '#/hooks/form'
import { ApiError } from '#/lib/http'
import type { DeposerAnnonceValues } from '#/features/annonce/schemas'

export const Route = createFileRoute('/annonces/deposer')({
  head: () => ({ meta: [{ title: 'Déposer une annonce' }] }),
  component: DeposerAnnoncePage,
})

const valeursInitiales: DeposerAnnonceValues = {
  titre: '',
  description: '',
  prix: '',
  categorie: '',
  ville: '',
  codePostal: '',
  pseudo: '',
  email: '',
  telephone: '',
  motDePasse: '',
}

const champs = new Set<string>(Object.keys(valeursInitiales))

// An ApiError that names a field of the form is shown under that field;
// any other error is shown above the button.
const erreurDeChamp = (error: unknown) =>
  error instanceof ApiError && error.champ && champs.has(error.champ)
    ? {
        champ: error.champ as keyof DeposerAnnonceValues,
        message: error.message,
      }
    : undefined

function DeposerAnnoncePage() {
  const { gateways, queryClient } = Route.useRouteContext()
  const [deposee, setDeposee] = useState(false)
  const toasts = useToastManager()
  const depot = useMutation(
    deposerAnnonceMutationOptions(gateways.annonce, queryClient),
  )

  const form = useAppForm({
    defaultValues: valeursInitiales,
    validators: { onChange: deposerAnnonceSchema },
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
        await depot.mutateAsync(deposerAnnonceSchema.parse(value))
        toasts.add({
          title: 'Annonce déposée',
          data: { intent: 'success', isClosable: true },
        })
        setDeposee(true)
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
    depot.error && !erreurDeChamp(depot.error) ? depot.error.message : undefined

  if (deposee) {
    return <EnCoursDeModeration />
  }

  return (
    <main className="page-wrap px-lg pt-2xl pb-2xl">
      <section className="mx-auto flex max-w-sz-640 flex-col gap-xl">
        <h1 className="text-display-2">Déposer une annonce</h1>
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

          <fieldset className="flex flex-col gap-lg">
            <legend className="mb-md text-headline-2">Vos coordonnées</legend>
            <form.AppField name="pseudo">
              {(field) => (
                <field.TextField
                  label="Pseudo"
                  description="Seul votre pseudo est affiché sur l'annonce."
                  autoComplete="nickname"
                  isRequired
                />
              )}
            </form.AppField>
            <form.AppField name="email">
              {(field) => (
                <field.TextField
                  label="Email"
                  type="email"
                  inputMode="email"
                  autoComplete="email"
                  isRequired
                />
              )}
            </form.AppField>
            <form.AppField name="telephone">
              {(field) => (
                <field.TextField
                  label="Téléphone (facultatif)"
                  type="tel"
                  inputMode="tel"
                  autoComplete="tel"
                />
              )}
            </form.AppField>
          </fieldset>

          <fieldset className="flex flex-col gap-lg">
            <legend className="mb-md text-headline-2">
              Suppression de l'annonce
            </legend>
            <form.AppField name="motDePasse">
              {(field) => (
                <field.PasswordField
                  label="Mot de passe de l'annonce"
                  description="Au moins 8 caractères. Il vous sera demandé pour supprimer l'annonce : conservez-le."
                  autoComplete="new-password"
                  isRequired
                />
              )}
            </form.AppField>
          </fieldset>

          {erreurGenerale && (
            <div
              role="alert"
              className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
            >
              {erreurGenerale}
            </div>
          )}

          <form.AppForm>
            <form.SubscribeButton label="Publier l'annonce" />
          </form.AppForm>
        </form>
      </section>
    </main>
  )
}

// A new annonce is not online yet: it waits for the moderation.
function EnCoursDeModeration() {
  return (
    <main className="page-wrap px-lg pt-3xl pb-2xl">
      <section className="flex flex-col items-start gap-lg rounded-xl border-sm border-outline bg-surface px-xl py-3xl text-on-surface">
        <h1 className="text-display-2">
          Votre annonce est en cours de modération
        </h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Elle sera en ligne dès qu'elle aura été validée. Un email de
          confirmation vous a été envoyé.
        </p>
        <Button asChild>
          <Link to="/">Retour à l'accueil</Link>
        </Button>
      </section>
    </main>
  )
}
