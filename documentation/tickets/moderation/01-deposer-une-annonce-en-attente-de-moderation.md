# 01: Déposer une annonce en attente de modération

**What to build:** quand un vendeur dépose une annonce, elle n'est plus en ligne tout de suite : elle attend la modération. Elle n'apparaît ni dans la recherche ni sur sa page, l'e-mail de confirmation et la page de confirmation du dépôt disent qu'elle est « en cours de modération », et l'événement `AnnonceDeposee` porte de quoi alimenter la modération sans que les deux contextes s'importent.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Une annonce nouvellement créée a le statut `EnAttenteModeration`. Les statuts `Challengee` et `Refusee` existent aussi, avec leurs transitions dans le domaine (`Activer`, `Challenger` avec motif, `Refuser`) ; seule `Activee` est visible.
- [ ] Une annonce non activée répond `ErrAnnonceIntrouvable` à la consultation, sort de la recherche et ne peut pas être supprimée par le vendeur, sans rien laisser fuiter.
- [ ] Les annonces de démo restent `Activee`, pour que la recherche et les guides QA continuent de fonctionner.
- [ ] `AnnonceDeposee` expose des accesseurs en primitives : identifiant, titre, description, prix en centimes, code de catégorie, ville, code postal, pseudo, date de dépôt.
- [ ] L'e-mail de confirmation de dépôt a pour sujet « Votre annonce « {titre} » est en cours de modération », garde ses liens et ne promet plus que l'annonce est en ligne.
- [ ] Après un dépôt, la page affiche que l'annonce est en cours de modération.
- [ ] Les tests couvrent : le statut à la création, l'invisibilité (consultation, recherche), l'événement et ses accesseurs, le nouvel e-mail, la page. Les tests existants et le parcours e2e du vendeur sont mis à jour.
- [ ] Le glossaire et l'`AGENTS.md` du bounded context `annonce` sont à jour.
- [ ] `make check` est vert.
