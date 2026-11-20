import { useMutation, useQuery } from '@tanstack/react-query'
import { Link, createFileRoute } from '@tanstack/react-router'
import { Button } from '@spark-ui/components/button'
import { ChallengerForm } from '#/components/ChallengerForm'
import { PseudoModerateurForm } from '#/components/PseudoModerateurForm'
import { libelleDeCategorie } from '#/features/annonce/categories'
import { formatPrix } from '#/features/annonce/format'
import { formatDateHeure, libelleEvenement } from '#/features/moderation/format'
import { usePseudoModerateur } from '#/features/moderation/pseudo-moderateur'
import {
  annonceAModererQueryOptions,
  challengerAnnonceMutationOptions,
  refuserAnnonceMutationOptions,
  relacherVerrouMutationOptions,
  validerAnnonceMutationOptions,
  verrouillerAnnonceMutationOptions,
} from '#/features/moderation/queries'
import type {
  AnnonceAModerer,
  StatutModeration,
} from '#/features/moderation/moderation-gateway'

// A hidden page: no link leads to it and search engines are told to skip it.
export const Route = createFileRoute('/moderation/$id')({
  head: () => ({
    meta: [
      { title: 'Modération d’une annonce' },
      { name: 'robots', content: 'noindex, nofollow' },
    ],
  }),
  component: AnnonceAModererPage,
})

function AnnonceAModererPage() {
  const [pseudo, setPseudo] = usePseudoModerateur()

  return (
    <main className="page-wrap flex flex-col gap-xl px-lg pt-2xl pb-2xl">
      <Link to="/moderation" className="text-body-2 text-main">
        ← Retour à la file
      </Link>
      {pseudo === undefined ? null : pseudo === '' ? (
        <PseudoModerateurForm onSubmit={setPseudo} />
      ) : (
        <Detail moderateur={pseudo} />
      )}
    </main>
  )
}

function Detail({ moderateur }: { moderateur: string }) {
  const { gateways } = Route.useRouteContext()
  const { id } = Route.useParams()
  const { data, isPending, error } = useQuery(
    annonceAModererQueryOptions(gateways.moderation, id),
  )

  if (isPending) {
    return <p className="text-body-1">Chargement de l'annonce…</p>
  }
  if (error) {
    return <Alerte message={error.message} />
  }
  if (!data) {
    return (
      <section className="flex flex-col gap-md">
        <h1 className="text-display-2">Annonce introuvable</h1>
        <p className="text-body-1 text-on-surface/dim-1">
          Cette annonce n'est pas, ou plus, dans la file de modération.
        </p>
      </section>
    )
  }
  return <Contenu annonce={data} moderateur={moderateur} />
}

const DECISIONS: Partial<Record<StatutModeration, string>> = {
  validee: 'Décision : annonce validée, elle est en ligne.',
  refusee: 'Décision : annonce refusée.',
  challengee: 'Décision : annonce challengée, le vendeur doit la corriger.',
}

function Contenu({
  annonce,
  moderateur,
}: {
  annonce: AnnonceAModerer
  moderateur: string
}) {
  const { gateways, queryClient } = Route.useRouteContext()
  const verrouiller = useMutation(
    verrouillerAnnonceMutationOptions(gateways.moderation, queryClient),
  )
  const relacher = useMutation(
    relacherVerrouMutationOptions(gateways.moderation, queryClient),
  )
  const valider = useMutation(
    validerAnnonceMutationOptions(gateways.moderation, queryClient),
  )
  const refuser = useMutation(
    refuserAnnonceMutationOptions(gateways.moderation, queryClient),
  )
  const challenger = useMutation(
    challengerAnnonceMutationOptions(gateways.moderation, queryClient),
  )
  const erreur =
    verrouiller.error ??
    relacher.error ??
    valider.error ??
    refuser.error ??
    challenger.error
  const decision = DECISIONS[annonce.statut]
  const aMoi = !decision && annonce.verrouillePar === moderateur
  const enCours =
    verrouiller.isPending ||
    relacher.isPending ||
    valider.isPending ||
    refuser.isPending ||
    challenger.isPending

  const agir = (action: () => void) => {
    verrouiller.reset()
    relacher.reset()
    valider.reset()
    refuser.reset()
    challenger.reset()
    action()
  }

  return (
    <article className="flex flex-col gap-xl">
      <header className="flex flex-col gap-sm">
        <h1 className="text-display-2">{annonce.titre}</h1>
        <p className="text-headline-2 text-main">{formatPrix(annonce.prix)}</p>
        <p className="text-body-2 text-on-surface/dim-1">
          {libelleDeCategorie(annonce.categorie) ?? annonce.categorie} ·{' '}
          {annonce.ville} ({annonce.codePostal}) · par {annonce.pseudo} ·
          déposée le {formatDateHeure(annonce.deposeeLe)}
        </p>
      </header>

      <p className="text-body-1 whitespace-pre-line">{annonce.description}</p>

      <section className="flex flex-col gap-md">
        {decision ? (
          <p className="text-body-1-highlight">{decision}</p>
        ) : (
          <>
            <p className="text-body-1-highlight">
              {annonce.verrouillePar
                ? `Verrouillée par ${annonce.verrouillePar}`
                : 'Libre'}
            </p>
            <div className="flex flex-wrap gap-md">
              {!aMoi && (
                <Button
                  disabled={enCours}
                  onClick={() =>
                    agir(() =>
                      verrouiller.mutate({ id: annonce.id, moderateur }),
                    )
                  }
                >
                  Verrouiller
                </Button>
              )}
              {aMoi && (
                <>
                  <Button
                    disabled={enCours}
                    onClick={() =>
                      agir(() => valider.mutate({ id: annonce.id, moderateur }))
                    }
                  >
                    Valider
                  </Button>
                  <Button
                    intent="danger"
                    disabled={enCours}
                    onClick={() =>
                      agir(() => refuser.mutate({ id: annonce.id, moderateur }))
                    }
                  >
                    Refuser
                  </Button>
                  <Button
                    design="outlined"
                    disabled={enCours}
                    onClick={() =>
                      agir(() =>
                        relacher.mutate({ id: annonce.id, moderateur }),
                      )
                    }
                  >
                    Relâcher
                  </Button>
                </>
              )}
            </div>
          </>
        )}
        {erreur && <Alerte message={erreur.message} />}
      </section>

      {aMoi && (
        <section className="flex flex-col gap-md">
          <h2 className="text-headline-1">Challenger l'annonce</h2>
          <ChallengerForm
            onSubmit={(motif) =>
              agir(() =>
                challenger.mutate({ id: annonce.id, moderateur, motif }),
              )
            }
          />
        </section>
      )}

      <section className="flex flex-col gap-md">
        <h2 className="text-headline-1">Historique</h2>
        <ul className="flex flex-col gap-sm">
          {annonce.historique.map((evenement, i) => (
            <li key={i} className="text-body-2">
              {formatDateHeure(evenement.date)} ·{' '}
              {libelleEvenement(
                evenement.type,
                evenement.moderateur,
                evenement.motif,
              )}
            </li>
          ))}
        </ul>
      </section>
    </article>
  )
}

function Alerte({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
    >
      {message}
    </div>
  )
}
