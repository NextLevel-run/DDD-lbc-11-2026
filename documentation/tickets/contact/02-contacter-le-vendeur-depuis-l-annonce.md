# 02: Contacter le vendeur depuis l'annonce

**What to build:** sur la page d'une annonce, l'acheteur clique sur « Contacter le vendeur ». Un formulaire de contact se déplie sous l'annonce, avec trois champs : son pseudo, son email et son message. À la soumission, le vendeur reçoit un email de la plateforme qui contient le message de l'acheteur. En cliquant sur « Répondre », il écrit directement à l'acheteur. L'adresse du vendeur n'est jamais montrée à l'acheteur. Le formulaire est alors remplacé par une confirmation. Si l'email n'a pas pu partir, l'acheteur le voit et peut réessayer sans ressaisir son message. Rien n'est conservé : le contact est un email plus un événement.

**Blocked by:** 01 — Passer le port Mailer à un email avec Reply-To

**Status:** done

- [x] La commande **Contacter** (`ContacterCommand`) est exposée par `POST /annonces/{id}/contact {pseudo, email, message}` et répond `204` en cas de succès.
- [x] Les règles de validation sont les suivantes : `pseudo` de 2 à 30 caractères, comme celui du vendeur ; `email` réduit à une adresse nue, comme celle du vendeur ; `message` de 10 à 2 000 caractères une fois trimé. Une saisie invalide répond `400 {"erreur", "champ"}` sur le champ en faute, et chaque nouvelle erreur a sa ligne dans la table des champs invalides du handler.
- [x] Seule une annonce activée peut être contactée. Une annonce inconnue ou supprimée répond `ErrAnnonceIntrouvable` (404), et un ID mal formé `ErrIDAnnonceInvalide` (400).
- [x] Le vendeur reçoit un email dont l'expéditeur est l'adresse fixe de la plateforme et dont le `Reply-To` est l'email de l'acheteur. L'objet est « Nouveau message pour votre annonce « {titre} » ». Le corps contient « {pseudo} vous a écrit : », puis le message, puis « Répondez directement à ce mail pour lui écrire. » et le lien `{FRONTEND_URL}/annonces/{id}`. L'email de l'acheteur ne figure pas dans le corps.
- [x] L'email est envoyé de façon synchrone par la commande. Si l'envoi échoue, elle renvoie `ErrEnvoiImpossible`, une erreur applicative que le handler traduit en `502 {"erreur": "Votre message n'a pas pu être envoyé, réessayez plus tard."}`, et rien n'est publié.
- [x] Après un envoi réussi, la commande publie **Formulaire de contact soumis** (`FormulaireContactSoumisEvent`, type `"FormulaireContactSoumis"`). L'événement porte l'ID de l'annonce, le pseudo, l'email et le message de l'acheteur, et `SoumisLe`, la date donnée par l'horloge injectée. Il ne porte jamais l'email ni le téléphone du vendeur.
- [x] Sur `/annonces/$id`, le bouton « Contacter le vendeur » déplie le formulaire (composants Spark). Le schéma Zod reprend les règles de l'API, et les erreurs du serveur s'affichent sous le champ en faute.
- [x] Après un envoi réussi, le formulaire est remplacé par « Votre message a été envoyé au vendeur. », sans réafficher le bouton. Après un 502, le message d'erreur s'affiche au-dessus du formulaire, qui garde la saisie.
- [x] Les tests couvrent les deux côtés, sur les adapters en mémoire :
  - côté API : validation, annonce introuvable, contenu et `Reply-To` du mail sur le spy, échec d'envoi sans événement, événement publié sur le collecteur ;
  - côté web : dépliage, validation, confirmation, erreur 502, sur la gateway en mémoire.
- [x] **Formulaire de contact soumis** a déjà sa ligne dans `domain.md` : le code reprend ce nom tel quel. L'`AGENTS.md` du bounded context `annonce` ne dit plus qu'il « ne sait rien de la messagerie » et décrit le flow Contacter.
- [x] Hors périmètre : la relance du vendeur à 7 jours, le KPI modé, la publicité, la copie du message à l'acheteur, et toute limitation d'envoi (anti-abus).
- [x] `make check` est vert.
