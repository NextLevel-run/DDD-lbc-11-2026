# annonce

The core bounded context: the annonces vendeurs put up for sale and acheteurs consult. It knows nothing of accounts or messaging, and moderation lives in its own context; they react to its events.

## Language

Shared terms: [`domain.md`](../../../../../documentation/glossary/domain.md), [`technical.md`](../../../../../documentation/glossary/technical.md).

- `NewAnnonceArgs` — the primitives an `Annonce` is built from, shared by `NewAnnonce` and `RestoreAnnonce`.
- `MotDePasse` — value object: pbkdf2-sha256 hash with a random salt, `Correspond(saisie)`; no getter, the clear text and the hash never leave the domain.
- `RaisonSuppression` — value object: closed list of codes (`vendue-sur-le-site`, `vendue-ailleurs`, `plus-envie-de-vendre`, `autre`), carried by `AnnonceSupprimee` only, never kept on the annonce.
- `RestoreAnnonce` — rebuilds an annonce whose ID and `Statut` are known (persistence, tests); `NewAnnonce` always makes a fresh UUID, `EnAttenteModeration`.
- `Statut` — `EnAttenteModeration` (new), `Activee` (the only visible one), `Challengee`, `Refusee`, `Supprimee`. `Activer`, `Challenger(motif)`, `Refuser` move an `EnAttenteModeration` annonce only (`ErrTransitionStatutInvalide` otherwise); the motif is kept on the annonce (`Motif()`).
- `CorrectionArgs` — the fields a vendeur may change (titre, description, prix, catégorie, ville, code postal), validated by `validerChamps`, the same rules as `RestoreAnnonce`. `AutoriserCorrection(saisie)` is the gate of `Corriger` and of the edition.

## Boundaries

- Driving: the HTTP handler (`GET /annonces`, `GET /annonces/{id}`, `POST /annonces`, `POST /annonces/{id}/suppression`, `POST /annonces/{id}/edition`, `POST /annonces/{id}/correction`, `GET /regions`, mounted by `RegisterRoutes`); the consumers in `adapter/driving/consumer` (`SubscribeConfirmationDepot`, and `SubscribeAnnonceValidee`, `SubscribeAnnonceChallengee`, `SubscribeAnnonceRefusee` for the decisions of moderation).
- Events produced: the contracts `AnnonceDeposeeEvent` and `AnnonceCorrigeeEvent` (an `InstantaneAnnonce` snapshot, built in `application/command/contract_events.go`), on the async bus of `main.go`; moderation queues both. `AnnonceSupprimeeEvent` (ID, `Raison`, `SupprimeeLe`) stays in `domain/` and has no consumer yet.
- Events consumed: its own `AnnonceDeposee` (confirmation de dépôt through the mailer port), and the contracts `AnnonceValideeEvent`, `AnnonceChallengeeEvent` (with `Motif`), `AnnonceRefuseeEvent` from moderation; the consumers type-assert them.
- Demo annonces are seeded by the composition root (`cmd/api/demo_annonces.go`), not by this context.

## Flows

- Déposer annonce: `NewAnnonce` (dated by the injected clock, `EnAttenteModeration`: invisible, undeletable until moderated) → `Save` → publish `AnnonceDeposee` → the ID, answered `201 {"id"}`. A validation error is a 400 `{"erreur", "champ"}`, `champ` being the JSON key at fault (table `invalidFields` in the handler); `motDePasse` (8+ characters, `ErrMotDePasseTropCourt`, 400 on `champ` `motDePasse`) is required and only hashed; a missing `prix` is refused by the handler, not taken as a don.
- Confirmation de dépôt: `AnnonceDeposee` → consumer → `EnvoyerConfirmationDepotCommand(annonceID)` → `GetByID` (the event carries no email) → mailer: subject « Votre annonce « {titre} » est en cours de modération », no promise that it is online, links `{FRONTEND_URL}/annonces/{id}` and `/supprimer` (default `http://localhost:3000`), fixed sender. A send failure is logged, never retried, and never fails the dépôt.
- Décisions de modération: `AnnonceValidee` → `ActiverAnnonceCommand` (statut `Activee`, visible) → `EnvoyerAnnonceValideeCommand`, subject « Votre annonce « {titre} » est validée », link `{FRONTEND_URL}/annonces/{id}`. `AnnonceChallengee` → `MarquerAnnonceChallengeeCommand` (statut `Challengee`, motif kept) → `EnvoyerAnnonceChallengeeCommand`, email with the motif and `{FRONTEND_URL}/annonces/{id}/corriger`. `AnnonceRefusee` → `MarquerAnnonceRefuseeCommand` (statut `Refusee`), **no email** (fraud). A failure (unknown annonce, bad transition, send) is logged, never retried, and never fails the decision. These commands publish nothing.
- Corriger annonce: `POST /annonces/{id}/edition {motDePasse}` → `ConsulterAnnonceAEditerQuery` → `200 {titre, description, prix, categorie, ville, codePostal, motif}` (prefills the form). `POST /annonces/{id}/correction {motDePasse, …fields}` → `ParseAnnonceID` → `GetByID` → `Annonce.Corriger` (authorise, validate, replace the fields, clear the motif, back to `EnAttenteModeration`; vendeur, mot de passe and date de dépôt stay) → `Save` → publish `AnnonceCorrigee` → `204`. Same 400s on `champ` as the dépôt (a missing `prix` is `prix`), wrong mot de passe 400 on `motDePasse`, malformed ID `ErrIDAnnonceInvalide`; unknown or not-`Challengee` annonce is `ErrAnnonceIntrouvable` (404), checked before the mot de passe.
- Rechercher annonces: `GET /annonces?motCle=&categorie=&prixMin=&prixMax=&region=&departement=&villeOuCodePostal=&page=` → `NewCriteresRecherche` + `NewPage` → `AnnonceRepository.Rechercher` (filter with `CriteresRecherche.Retient`, newest `deposeeLe` first, `ParPage` = 20) → `{"annonces", "total", "page", "parPage"}`. A page past the last is an empty list, not an error; a negative or non-integer page is a 400 on `champ` `page`. Prix bounds are integer centimes, both included (0–0 finds the dons); a negative or non-integer bound is a 400 on `prixMin`/`prixMax`, a minimum above the maximum a 400 on `prixMin`, an unknown catégorie a 400 on `categorie`. An unknown région or département is a 400 on `region`/`departement`, both at once a 400 on `region` (`ErrLieuAmbigu`); an incoherent lieu (Ajaccio in 93) is zero results, no error.
- Supprimer annonce: `POST /annonces/{id}/suppression {motDePasse, raison}` → `ParseAnnonceID` → `NewRaisonSuppression` → `GetByID` → `Annonce.Supprimer` (statut `Supprimee`, stays in the repository) → `Save` → publish `AnnonceSupprimee` → `204`. Wrong mot de passe: `ErrMotDePasseIncorrect`, 400 on `champ` `motDePasse`; missing or unknown raison: 400 on `raison`; unknown or already supprimee: `ErrAnnonceIntrouvable` (404). Attempts are not limited.
- Lister régions: `GET /regions` → `[{"code", "nom", "departements": [{"code", "nom"}]}]`, from the static table; feeds the lieu selector.
- Consulter annonce: `ParseAnnonceID` → `GetByID` → `EstActivee` → `AnnonceView`. A malformed ID is `ErrIDAnnonceInvalide` (400); unknown or not activee is `ErrAnnonceIntrouvable` (404).

## Invariants and pitfalls

- **A view never exposes the vendeur's email or telephone**: only the pseudo leaves the domain.
- **The mot de passe never appears** in a view, an event, a log or a JSON response; `Annonce.MotDePasseCorrespond` is the only question asked. Demo annonces all use `demo1234`.
- **No email contains the mot de passe**, only the sentence that the suppression (or the correction) will ask for it.
- **A failed correction changes nothing**: `Corriger` validates every field before assigning any.
- **A non-activee annonce answers `ErrAnnonceIntrouvable`**, never a distinct error: its existence must not leak. It holds for the consultation, the search (`Retient`) and the suppression; the edition and the correction answer it for any annonce that is not `Challengee`, whatever the mot de passe. Demo annonces call `Activer` before `Save`.
- **Prix is in centimes**, zero is a don, no ceiling. The frontend converts to euros.
- **Catégorie codes are stable** (`vehicules`, `maison-jardin`…): they travel in JSON and URLs; only the libellé is French prose.
- **Département from the code postal**: two-digit prefix 01–95; Corse splits on the third digit (200–201 → 2A, 202–206 → 2B, 207–209 rejected); overseas reads three digits, 971–974 and 976. 975 (Saint-Pierre-et-Miquelon) is no département and is rejected, as are 00, 96, 98, 99.
- **Région and département codes are INSEE codes**; the full table is `domain/departement.go`.
- **Validation lives in the constructors**: titre 3–100 and description 10–4 000 characters (runes, trimmed), pseudo 2–30, bare email, telephone `0` + 9 digits once spaces, dots and dashes are dropped. The web schema mirrors these rules: change both together.
- **A new domain error needs its line in `invalidFields`**, or the form gets a 500 instead of a message under the field.
- **`CriteresRecherche` holds every search criterion**: a new filter joins `CriteresRechercheArgs` and `Retient`, not a new repository method. `Retient` also enforces "activee only".
- **`villeOuCodePostal`**: five digits (trimmed) are compared exactly to the code postal, never as a prefix; anything else must equal the normalised ville as a whole (« Paris » does not find « Paris 15e »).
- **`ErrCodeDepartementInconnu` is not `ErrDepartementInconnu`**: the first is an unknown code in a recherche, the second a code postal that points to no département in a dépôt.
- **Text comparison goes through `normaliser`** (`domain/normalisation.go`): lower case, accents dropped by a table (no `x/text`), hyphens and apostrophes as spaces. A mot-clé word matches as a substring, so `elec` finds « électrique ».
- **The list view is `AnnonceResumeView`**: titre, prix, ville, images, date; no vendeur at all.
- **JSON keys are French** and errors answer `{"erreur": "…"}` in French.
