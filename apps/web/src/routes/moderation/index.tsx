import { useQuery } from '@tanstack/react-query'
import { Link, createFileRoute } from '@tanstack/react-router'
import { PseudoModerateurForm } from '#/components/PseudoModerateurForm'
import { formatPrix } from '#/features/annonce/format'
import { formatDateHeure } from '#/features/moderation/format'
import { usePseudoModerateur } from '#/features/moderation/pseudo-moderateur'
import { annoncesAModererQueryOptions } from '#/features/moderation/queries'

// A hidden page: no link leads to it and search engines are told to skip it.
export const Route = createFileRoute('/moderation/')({
  head: () => ({
    meta: [
      { title: 'Modération' },
      { name: 'robots', content: 'noindex, nofollow' },
    ],
  }),
  component: FileDeModeration,
})

function FileDeModeration() {
  const [pseudo, setPseudo] = usePseudoModerateur()

  return (
    <main className="page-wrap flex flex-col gap-xl px-lg pt-2xl pb-2xl">
      <h1 className="text-display-2">Modération</h1>
      {pseudo === undefined ? null : pseudo === '' ? (
        <PseudoModerateurForm onSubmit={setPseudo} />
      ) : (
        <>
          <p className="text-body-1 text-on-surface/dim-1">
            Connecté en tant que {pseudo}.
          </p>
          <Liste />
        </>
      )}
    </main>
  )
}

function Liste() {
  const { gateways } = Route.useRouteContext()
  const { data, isPending, error } = useQuery(
    annoncesAModererQueryOptions(gateways.moderation),
  )

  if (isPending) {
    return <p className="text-body-1">Chargement de la file…</p>
  }
  if (error) {
    return (
      <div
        role="alert"
        className="rounded-md bg-error-container p-md text-body-2 text-on-error-container"
      >
        {error.message}
      </div>
    )
  }
  if (data.length === 0) {
    return <p className="text-body-1">Aucune annonce à modérer.</p>
  }
  return (
    <ul className="flex flex-col gap-md">
      {data.map((annonce) => (
        <li
          key={annonce.id}
          className="flex flex-col gap-sm rounded-xl border-sm border-outline bg-surface p-lg text-on-surface"
        >
          <Link
            to="/moderation/$id"
            params={{ id: annonce.id }}
            className="text-headline-2 text-main"
          >
            {annonce.titre}
          </Link>
          <p className="text-body-2 text-on-surface/dim-1">
            {formatPrix(annonce.prix)} · {annonce.ville} · par {annonce.pseudo}{' '}
            · en file depuis le {formatDateHeure(annonce.misEnFileLe)}
          </p>
          {annonce.verrouillePar && (
            <p className="text-body-2-highlight">
              Verrouillée par {annonce.verrouillePar}
            </p>
          )}
        </li>
      ))}
    </ul>
  )
}
