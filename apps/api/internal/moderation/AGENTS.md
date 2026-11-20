# moderation

The bounded context that holds the annonces déposées in a file de modération, lets a modérateur lock, release and decide them (valider, refuser, challenger). It never imports `annonce`: it reads `AnnonceDeposee` and `AnnonceCorrigee` through a local interface, and annonce reads its decisions the same way.

## Language

Shared terms: [`domain.md`](../../../../documentation/glossary/domain.md), [`technical.md`](../../../../documentation/glossary/technical.md).

- `HistoriqueAnnonce` — the event-sourced aggregate: one stream per annonce, ID = the annonce's ID. `RestoreHistoriqueAnnonce(id, events)` replays; a decision checks the replayed state, then appends; `NouveauxEvenements()` is what the store appends; `Version()` is the count of events already stored.
- `InstantaneAnnonce` — primitives copied from the deposit, carried by `AnnonceMiseEnFile`; the only annonce data moderation holds.
- `Moderateur` — value object, a pseudo of 2 to 30 characters once trimmed; no account, no aggregate.
- `Motif` — value object, the reason of a challenge: 10 to 1 000 characters once trimmed.
- `Etat` — `EtatVide` (no stream), `EtatEnFile`, `EtatVerrouillee`, then the decided `EtatValidee`, `EtatRefusee` (out of the queue for good) and `EtatChallengee` (out until the vendeur corrects).

## Boundaries

- Driving: the HTTP handler (`GET /moderation/annonces`, `GET /moderation/annonces/{id}`, `POST …/verrou`, `…/relachement`, `…/validation`, `…/refus` {moderateur}, `…/challenge` {moderateur, motif}; no authentication), mounted by `RegisterRoutes`; the consumer `SubscribeMiseEnFile` on `AnnonceDeposee` and `AnnonceCorrigee`.
- Driven: `HistoriqueAnnonceRepository`, the event store (`Charger`, `Enregistrer`, `Tous`), in memory.
- Events consumed: `AnnonceDeposee` and `AnnonceCorrigee` (same accessors, one handler), through `interface{ AnnonceID() string … }` declared in the consumer. Events published: `AnnonceValidee`, `AnnonceRefusee`, `AnnonceChallengee` (accessors `AnnonceID()`, and `Motif()` for the last), after the append; lock events stay in the stream.

## Flows

- Mettre en file: `AnnonceDeposee` or `AnnonceCorrigee` → `MettreEnFileCommand` → `Charger` → `MettreEnFile` → `Enregistrer`. An empty stream or a `Challengee` one (the correction: new `AnnonceMiseEnFile` with the corrected snapshot, history kept) goes to `EnFile`; any other state is ignored, never an error.
- Verrouiller: `NewModerateur` → `Charger` → `Verrouiller(moderateur, now)` → `Enregistrer` → `204`. Lock states: `EnFile` → `Verrouillee` (`AnnonceVerrouillee`) → `EnFile` (`VerrouAnnonceRelache`, cause `manuel` or `expiration`).
- Relâcher: same shape with `Relacher`; only the holder of a valid lock may.
- Valider / refuser / challenger: `NewModerateur` (and `NewMotif`) → `Charger` → decision method (needs a valid lock of this moderator, else `ErrAnnoncePasVerrouilleeParModerateur` or `ErrVerrouExpire`, 409; it returns the event it appended) → `Enregistrer` → publish that event → `204`. An invalid motif is `ErrMotifInvalide`, 400 on `motif`, checked before the stream is loaded.
- Lister / consulter: replay `Tous()` or one stream; a decided annonce is not listed and shows its decision as `statut` (`validee`, `refusee`, `challengee`); a lock older than 2 hours is shown as free (the stream keeps its trace until the next lock).

## Invariants and pitfalls

- **The clock is injected** (`now func() time.Time`); the aggregate never reads `time.Now`. A lock lasts `DureeVerrou` (2 h): valid at exactly 2 h, expired after.
- **Expiry is lazy**: nothing runs on a timer. `Verrouiller` appends `VerrouAnnonceRelache{expiration}` then `AnnonceVerrouillee`; `Relacher` on an expired lock is `ErrVerrouExpire`, from another moderator `ErrAnnoncePasVerrouilleeParModerateur`.
- **The holder locking again is idempotent**: no event, no error.
- **An unknown annonce is `ErrAnnonceIntrouvable` (404)**, including a malformed ID: the empty stream is the only test. Lock and state errors are 409, `ErrModerateurInvalide` 400 on `moderateur`.
- **A decision ends the flow, except a challenge**: only `AnnonceCorrigee` puts a `Challengee` annonce back in the queue; a `Validee` or `Refusee` one ignores it. The annonce is no longer lockable (`ErrAnnoncePasEnFile`) nor decidable again (`ErrAnnoncePasVerrouilleeParModerateur`). The decision events are stored and published: the same value, unexported `annonceID`/`motif` behind accessors.
- **The store checks the version**: `Enregistrer` of a stale aggregate is `ErrConflitVersion` (409); the aggregate is not reusable after an append, reload it.
- **A new domain error needs its line in `respondDomainError`**, or it answers a 500.
- **Stream event types are stable strings** (`AnnonceMiseEnFile`, `AnnonceVerrouillee`, `VerrouAnnonceRelache`, `AnnonceValidee`, `AnnonceRefusee`, `AnnonceChallengee`): they are the storage format.
