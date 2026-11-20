# 02: Voir et verrouiller les annonces à modérer

**What to build:** un modérateur ouvre la page cachée `/moderation`, saisit son pseudo et voit la file des annonces déposées, les plus anciennes d'abord. Il ouvre l'une d'elles, en voit le contenu et l'historique, la verrouille pour la traiter ou relâche son verrou. Un verrou de plus de 2 heures est considéré comme libre. Le nouveau bounded context `moderation` repose sur l'event sourcing.

**Blocked by:** 01 (Déposer une annonce en attente de modération)

**Status:** ready-for-agent

- [ ] Le bounded context `moderation` est créé (carte du domaine, `AGENTS.md`). Il n'importe jamais `annonce` : il consomme `AnnonceDeposee` par une interface locale.
- [ ] `HistoriqueAnnonce` est un agrégat en event sourcing : un flux d'événements par annonce, rejoué pour décider, les nouveaux événements ajoutés avec un contrôle de version (`ErrConflitVersion`). Le port d'event store a un adaptateur en mémoire.
- [ ] `AnnonceMiseEnFile` porte l'instantané de l'annonce. La file est une lecture calculée en rejouant les flux, pas un agrégat.
- [ ] Verrouiller : possible depuis la file ; verrou valide d'un autre modérateur : `ErrAnnonceDejaVerrouillee` (409) ; même modérateur : idempotent ; verrou de plus de 2 h : `VerrouAnnonceRelache` (cause `expiration`) puis nouveau verrou, évalués avec l'horloge injectée.
- [ ] Relâcher : réservé au détenteur d'un verrou valide, remet l'annonce en file ; sinon `ErrAnnoncePasVerrouilleeParModerateur` ou `ErrVerrouExpire` (409).
- [ ] Le pseudo du modérateur fait 2 à 30 caractères (`ErrModerateurInvalide`, 400 sur `moderateur`).
- [ ] L'API expose `GET /moderation/annonces`, `GET /moderation/annonces/{id}`, `POST …/verrou` et `POST …/relachement`, tels que décrits dans la spec ; une annonce inconnue est un 404.
- [ ] Les pages `/moderation` (pseudo conservé, liste avec « Verrouillée par … ») et `/moderation/$id` (contenu, historique, Verrouiller, Relâcher) n'ont aucun lien dans la navigation et sont en `noindex`. Les erreurs 409 s'affichent en français.
- [ ] Le contexte est branché dans le composition root et couvert par `server_test.go`. Les tests couvrent le domaine (rejeu, verrou, expiration), les commandes avec le collecteur d'événements, les requêtes, le consommateur et les pages sur la gateway en mémoire.
- [ ] Les glossaires, l'`AGENTS.md` de `moderation` et ceux de l'API et du web sont à jour.
- [ ] `make check` est vert.
