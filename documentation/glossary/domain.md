# Glossaire du domaine

Langage ubiquitaire, en français ; le code reprend ces mots sans accents et sans les traduire. Une
ligne par racine d'agrégat, acteur ou événement : ce qu'il **est**. Ni le détail d'un agrégat, ni
les commandes et requêtes : le code les nomme. Format : `- **Terme** (`NomEnCode`) — définition.`,
la parenthèse seulement si le nom diffère.

## Carte du domaine

| Sous-domaine | Type | Bounded context | Agrégats |
| ------------ | ---- | --------------- | -------- |
| Annonce      | cœur | `annonce`       | Annonce  |
| Modération   | support | `moderation` | HistoriqueAnnonce |

## Annonce

- **Annonce** — un bien qu'un vendeur met en vente.
- **Vendeur** — la personne qui dépose une annonce, sans compte.
- **Acheteur** — la personne qui recherche et consulte des annonces.
- **Annonce déposée** (`AnnonceDeposee`) — l'événement d'un dépôt ; il déclenche l'email de confirmation de dépôt et alimentera la modération.
- **En attente de modération** (`EnAttenteModeration`) — statut d'une annonce déposée : invisible partout, jusqu'à sa validation.
- **Annonce challengée** (`Challengee`) — la modération demande une correction au vendeur, avec un motif ; l'annonce reste invisible.
- **Annonce corrigée** (`AnnonceCorrigee`) — le vendeur a corrigé son annonce challengée avec son mot de passe ; elle repasse en attente de modération.
- **Annonce refusée** (`Refusee`) — la modération a rejeté l'annonce ; invisible, définitivement.
- **Mot de passe de l'annonce** (`MotDePasse`) — choisi par le vendeur au dépôt (8 caractères minimum), conservé haché ; il autorise suppression et correction.
- **Annonce supprimée** (`AnnonceSupprimee`) — le vendeur a retiré son annonce avec son mot de passe ; statut `Supprimee`, elle n'est plus en ligne.
- **Raison de suppression** (`RaisonSuppression`) — pourquoi le vendeur retire son annonce, un code d'une liste fermée ; seul l'événement la porte.

## Modération

- **Modérateur** — la personne qui traite la file ; sans compte, elle saisit un pseudo de 2 à 30 caractères.
- **Historique d'annonce** (`HistoriqueAnnonce`) — le flux d'événements d'une annonce en modération ; son état se déduit en le rejouant.
- **File de modération** — les annonces déposées à traiter, les plus anciennes d'abord ; une lecture, pas un agrégat.
- **Annonce mise en file** (`AnnonceMiseEnFile`) — l'événement d'entrée dans la file ; il porte l'instantané de l'annonce.
- **Annonce verrouillée** (`AnnonceVerrouillee`) — un modérateur s'en est saisi ; au-delà de 2 heures le verrou est libre.
- **Verrou relâché** (`VerrouAnnonceRelache`) — le verrou a pris fin, à la main (`manuel`) ou par `expiration` ; l'annonce retourne en file.
- **Annonce validée** (`AnnonceValidee`) — le modérateur qui la détient la met en ligne ; le vendeur en est prévenu par email.
- **Annonce refusée** (`AnnonceRefusee`) — décision du modérateur : l'annonce reste invisible, sans email au vendeur (fraude).
- **Annonce challengée** (`AnnonceChallengee`) — décision du modérateur : le vendeur doit corriger ; son email porte le motif.
- **Motif** (`Motif`) — pourquoi la modération challenge, de 10 à 1 000 caractères ; il est conservé sur l'annonce et envoyé au vendeur.
