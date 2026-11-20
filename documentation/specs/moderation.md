# Spec: modération des annonces

Source : [`prompts/moderation.md`](../../prompts/moderation.md) et le parcours 2 de l'[event storming](../event-storming/event-storming.md).

## Décisions (cadrage)

- **Périmètre** : parcours 2 complet (file, verrou, valider, refuser, challenger, relâcher, expiration du verrou à 2 h, correction par le vendeur). Le signalement (parcours 4), l'archivage, le KPI modé et l'entité Modérateur sont hors lot.
- **Event sourcing** sur l'agrégat **Historique d'annonce** : un flux d'événements par annonce, rejoué pour décider. La « file d'attente » n'est pas un agrégat : c'est une lecture calculée en rejouant les flux. Réponse à la question grise « deux agrégats ou un ? » : un seul.
- **Visibilité** : une annonce déposée est invisible (recherche, consultation) jusqu'à `AnnonceValidee`.
- **Page cachée** : `/moderation`, URL non liée, sans authentification. Le modérateur saisit un pseudo (2 à 30 caractères) ; pas d'agrégat Modérateur.
- **Verrou de 2 h** : expiration paresseuse, évaluée avec l'horloge injectée à la lecture et à l'action, sans tâche planifiée.
- **E-mails** : validée et challengée, par consommateurs dans `annonce`, via le port Mailer existant.

## Bounded contexts

- `moderation` (nouveau, sous-domaine Modération) : agrégat `HistoriqueAnnonce` en event sourcing. N'importe jamais `annonce`.
- `annonce` : nouveaux statuts, correction, consommateurs des décisions.

Les deux ne se parlent que par événements sur le bus. Un événement qui franchit la frontière expose des **accesseurs en primitives** ; le consommateur de l'autre contexte déclare sa propre interface locale (`interface{ AnnonceID() string … }`) et n'importe pas le type de l'émetteur.

## Contexte `annonce`

### Statuts

`EnAttenteModeration` (à la création : `NewAnnonce` ne produit plus `Activee`), `Activee`, `Challengee`, `Refusee`, `Supprimee`. Seule `Activee` est visible (recherche, consultation, suppression). Toute annonce non activée répond `ErrAnnonceIntrouvable` aux lectures publiques. `RestoreAnnonce` reste le moyen de charger un statut connu (les annonces de démo sont `Activee`).

### Événements

| Événement | Type | Accesseurs (primitives) | Émis par |
| --- | --- | --- | --- |
| `AnnonceDeposeeEvent` | `"AnnonceDeposee"` | `AnnonceID()`, `Titre()`, `Description()`, `Prix()` (centimes, `int64`), `Categorie()` (code), `Ville()`, `CodePostal()`, `Pseudo()`, `DeposeeLe()` | dépôt (existant, accesseurs ajoutés) |
| `AnnonceCorrigeeEvent` | `"AnnonceCorrigee"` | les mêmes | correction |

### Consommateurs (`adapter/driving/consumer`)

| Reçoit | Commande | Effet |
| --- | --- | --- |
| `AnnonceValidee` (`AnnonceID()`) | `ActiverAnnonceCommand` | statut `Activee`, e-mail « Votre annonce « {titre} » est validée » avec le lien `{FRONTEND_URL}/annonces/{id}` |
| `AnnonceChallengee` (`AnnonceID()`, `Motif()`) | `MarquerAnnonceChallengeeCommand` | statut `Challengee`, motif conservé sur l'annonce, e-mail avec le motif et le lien `{FRONTEND_URL}/annonces/{id}/corriger` |
| `AnnonceRefusee` (`AnnonceID()`) | `MarquerAnnonceRefuseeCommand` | statut `Refusee`, **aucun e-mail** (fraude) |

Un échec d'envoi est journalisé, jamais réessayé, et ne fait pas échouer la décision (comme la confirmation de dépôt). L'e-mail de confirmation de dépôt devient « Votre annonce « {titre} » est en cours de modération » (même liens, sans promettre qu'elle est en ligne).

### Correction par le vendeur

- `POST /annonces/{id}/edition` `{motDePasse}` → `200 {titre, description, prix, categorie, ville, codePostal, motif}` : de quoi préremplir le formulaire. Réservé aux annonces `Challengee`.
- `POST /annonces/{id}/correction` `{motDePasse, titre, description, prix, categorie, ville, codePostal}` → `204`. Mêmes validations que le dépôt (400 avec `champ`). `Annonce.Corriger` remplace ces champs, efface le motif, repasse en `EnAttenteModeration`, puis `AnnonceCorrigee` est publié après persistance.
- Erreurs : ID malformé `ErrIDAnnonceInvalide` (400) ; annonce inconnue **ou pas `Challengee`** : `ErrAnnonceIntrouvable` (404, rien ne fuite) ; mot de passe faux : `ErrMotDePasseIncorrect` (400, `champ` `motDePasse`). Vendeur (pseudo, e-mail, téléphone), mot de passe et date de dépôt ne changent pas.

## Contexte `moderation`

### Agrégat `HistoriqueAnnonce` (event sourcing)

Identité : l'ID de l'annonce. Reconstruit par `RestoreHistoriqueAnnonce(événements)` ; chaque méthode de décision vérifie l'état courant puis **ajoute** un événement (`NouveauxEvenements()`), sans muter un état persisté. Version = nombre d'événements déjà stockés.

Événements du flux (types stables) :

| Événement | Données |
| --- | --- |
| `AnnonceMiseEnFile` | instantané de l'annonce : `AnnonceID`, `Titre`, `Description`, `Prix`, `Categorie`, `Ville`, `CodePostal`, `Pseudo`, `DeposeeLe`, date |
| `AnnonceVerrouillee` | `Moderateur`, date |
| `VerrouAnnonceRelache` | `Moderateur` (vide si expiration), `Cause` (`manuel` ou `expiration`), date |
| `AnnonceValidee` | `Moderateur`, date |
| `AnnonceRefusee` | `Moderateur`, date |
| `AnnonceChallengee` | `Moderateur`, `Motif`, date |

États dérivés du rejeu : `EnFile`, `Verrouillee(moderateur, depuis)`, `Validee`, `Refusee`, `Challengee`.

Règles (horloge `now` passée en paramètre) :

- `MettreEnFile(instantané)` : ajoute `AnnonceMiseEnFile`. Valable pour un flux vide (dépôt) ou `Challengee` (correction : retour dans la file). Un flux `Refusee` ou `Validee` ignore silencieusement (l'événement n'est pas une erreur).
- `Verrouiller(moderateur, now)` : depuis `EnFile`. Si l'annonce est `Verrouillee` depuis plus de 2 h, ajoute d'abord `VerrouAnnonceRelache{expiration}` puis `AnnonceVerrouillee`. Verrou valide d'un autre : `ErrAnnonceDejaVerrouillee` (409). Déjà verrouillée par le même : idempotent (aucun événement). Autre état : `ErrAnnoncePasEnFile` (409).
- `Valider`, `Refuser`, `Challenger(motif)`, `Relacher` : exigent un verrou **valide** détenu par ce modérateur ; sinon `ErrAnnoncePasVerrouilleeParModerateur` (409), ou `ErrVerrouExpire` (409) si le verrou a plus de 2 h. `Challenger` exige un `motif` de 10 à 1 000 caractères une fois rogné (`ErrMotifInvalide`, 400, `champ` `motif`). `Relacher` ajoute `VerrouAnnonceRelache{manuel}` et remet l'annonce en file.
- `moderateur` : 2 à 30 caractères une fois rogné, sinon `ErrModerateurInvalide` (400, `champ` `moderateur`).

### Ports

`HistoriqueAnnonceRepository` (event store) : `Charger(annonceID) (*HistoriqueAnnonce, error)` (flux vide = historique vide, pas d'erreur) ; `Enregistrer(h *HistoriqueAnnonce) error` ajoute les nouveaux événements avec contrôle de version optimiste (`ErrConflitVersion`) ; `Tous() ([]*HistoriqueAnnonce, error)` pour les lectures. L'adaptateur en mémoire garde un journal d'événements par annonce.

### Commandes, requêtes, événements publiés

- Commandes : `MettreEnFileCommand` (consommateur de `AnnonceDeposee` et `AnnonceCorrigee`), `VerrouillerAnnonceCommand`, `RelacherVerrouCommand`, `ValiderAnnonceCommand`, `RefuserAnnonceCommand`, `ChallengerAnnonceCommand`. Chacune charge, appelle l'agrégat, enregistre, puis publie sur le bus (après persistance) : `AnnonceValidee`, `AnnonceRefusee`, `AnnonceChallengee` uniquement, avec les accesseurs `AnnonceID()` (et `Motif()`). Les autres événements restent dans le flux.
- Requêtes : `ListerAnnoncesAModererQuery` (annonces `EnFile` et `Verrouillee`, verrous expirés comptés comme libres, plus anciennes d'abord) ; `ConsulterAnnonceAModererQuery` (instantané, état, verrou, historique daté des événements). Vues : `AnnonceAModererResumeView`, `AnnonceAModererView`.
- Consommateurs : `adapter/driving/consumer` souscrit à `AnnonceDeposee` et `AnnonceCorrigee` par interface locale.

### HTTP (`RegisterRoutes`, aucune authentification)

| Route | Corps | Réponse |
| --- | --- | --- |
| `GET /moderation/annonces` | | `200 {"annonces":[{id, titre, prix, ville, pseudo, misEnFileLe, verrouillePar, verrouilleDepuis}]}` (`verrouillePar` vide si libre) |
| `GET /moderation/annonces/{id}` | | `200 {id, titre, description, prix, categorie, ville, codePostal, pseudo, deposeeLe, statut, verrouillePar, verrouilleDepuis, historique:[{type, date, moderateur, motif}]}` ; inconnue : 404 |
| `POST /moderation/annonces/{id}/verrou` | `{moderateur}` | `204` |
| `POST /moderation/annonces/{id}/relachement` | `{moderateur}` | `204` |
| `POST /moderation/annonces/{id}/validation` | `{moderateur}` | `204` |
| `POST /moderation/annonces/{id}/refus` | `{moderateur}` | `204` |
| `POST /moderation/annonces/{id}/challenge` | `{moderateur, motif}` | `204` |

Statuts : 400 validation (avec `champ`), 404 `ErrAnnonceIntrouvable`, 409 pour les erreurs de verrou et d'état. `statut` ∈ `en-file`, `verrouillee`, `validee`, `refusee`, `challengee`. Les clés JSON sont en français.

## Frontend

- Feature `moderation` (gateway port, HTTP gateway, in-memory gateway, query options, schémas) ; routes `/moderation` (liste) et `/moderation/$id` (détail et décision). Aucun lien depuis la navigation ; `noindex`.
- `/moderation` : saisie du pseudo modérateur (conservé dans `localStorage`), liste des annonces à modérer avec l'indication « Verrouillée par … ». `/moderation/$id` : contenu de l'annonce, historique, boutons Verrouiller / Valider / Refuser / Challenger (motif obligatoire) / Relâcher selon l'état et le verrou, erreurs 409 affichées en français.
- Annonce : la page `/annonces/$id/corriger` demande le mot de passe, préremplit le formulaire (avec le motif du challenge), envoie la correction et confirme « Votre annonce a été corrigée, elle repart en modération ». Après un dépôt, le message dit que l'annonce est en cours de modération. Composants Spark, comme les autres pages.
- Tests sur les gateways en mémoire, sans mock de `fetch`.

## Tranches verticales

1. Dépôt invisible : statut `EnAttenteModeration`, événements à accesseurs, e-mail de confirmation mis à jour, `Rechercher`/`Consulter` inchangés grâce à `EstActivee`.
2. File et verrou : contexte `moderation`, agrégat, event store, consommateur `AnnonceDeposee`, liste, détail, verrouiller, relâcher, expiration paresseuse, page `/moderation`.
3. Décider : valider (activation + e-mail), refuser, challenger (statut + e-mail).
4. Corriger : `edition`, `correction`, `AnnonceCorrigee`, retour en file, page `/annonces/$id/corriger`.
