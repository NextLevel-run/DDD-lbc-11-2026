# Glossaire du domaine

Langage ubiquitaire, en français ; le code reprend ces mots sans accents et sans les traduire. Une
ligne par racine d'agrégat, acteur ou événement : ce qu'il **est**. Ni le détail d'un agrégat, ni
les commandes et requêtes : le code les nomme. Format : `- **Terme** (`NomEnCode`) — définition.`,
la parenthèse seulement si le nom diffère.

## Carte du domaine

| Sous-domaine | Type | Bounded context | Agrégats |
| ------------ | ---- | --------------- | -------- |
| Annonce      | cœur | `annonce`       | Annonce  |

## Annonce

- **Annonce** — un bien qu'un vendeur met en vente.
- **Vendeur** — la personne qui dépose une annonce, sans compte.
- **Acheteur** — la personne qui recherche et consulte des annonces.
- **Annonce déposée** (`AnnonceDeposee`) — l'événement d'un dépôt ; il déclenche l'email de confirmation de dépôt, la modération s'y branchera.
- **Mot de passe de l'annonce** (`MotDePasse`) — choisi par le vendeur au dépôt (8 caractères minimum), conservé haché ; il autorisera la suppression.
- **Annonce supprimée** (`AnnonceSupprimee`) — le vendeur a retiré son annonce avec son mot de passe ; statut `Supprimee`, elle n'est plus en ligne.
- **Raison de suppression** (`RaisonSuppression`) — pourquoi le vendeur retire son annonce, un code d'une liste fermée ; seul l'événement la porte.
- **Formulaire de contact soumis** (`FormulaireContactSoumis`) — un acheteur a écrit au vendeur d'une annonce ; le message lui est parti par mail.
