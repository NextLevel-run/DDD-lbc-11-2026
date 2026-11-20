# 04: Corriger une annonce challengée

**What to build:** depuis le lien de l'e-mail, le vendeur d'une annonce challengée arrive sur la page de correction. Il saisit le mot de passe de l'annonce, voit le formulaire prérempli et le motif du challenge, corrige puis envoie. Une page lui confirme que l'annonce repart en modération, et elle retourne dans la file du modérateur.

**Blocked by:** 03 (Valider, refuser ou challenger une annonce)

**Status:** ready-for-agent

- [ ] `POST /annonces/{id}/edition` avec `{motDePasse}` renvoie les champs modifiables (titre, description, prix, catégorie, ville, code postal) et le motif. Il est réservé aux annonces `Challengee`.
- [ ] `POST /annonces/{id}/correction` avec le mot de passe et ces champs répond `204`, avec les mêmes validations et les mêmes erreurs 400 sur `champ` que le dépôt. `Annonce.Corriger` remplace les champs, efface le motif et repasse l'annonce en `EnAttenteModeration` ; vendeur, mot de passe et date de dépôt ne changent pas.
- [ ] Mot de passe faux : `ErrMotDePasseIncorrect` (400 sur `motDePasse`). ID malformé : `ErrIDAnnonceInvalide`. Annonce inconnue ou pas `Challengee` : `ErrAnnonceIntrouvable` (404), sans rien laisser fuiter.
- [ ] `AnnonceCorrigee` est publié après persistance, avec les mêmes accesseurs que `AnnonceDeposee`. Le contexte `moderation` le consomme et ajoute `AnnonceMiseEnFile` au flux : l'annonce revient dans la file, avec son historique complet. Une annonce `Refusee` ou `Validee` ignore l'événement sans erreur.
- [ ] La page `/annonces/$id/corriger` utilise les composants Spark : mot de passe, puis formulaire prérempli avec le motif, puis la confirmation « Votre annonce a été corrigée, elle repart en modération ». Annonce introuvable : « Annonce introuvable ».
- [ ] Le parcours complet est couvert en e2e : dépôt, challenge par le modérateur, correction, nouvelle validation, annonce visible.
- [ ] Les tests couvrent le domaine, la commande, le consommateur, `server_test.go` et la page sur la gateway en mémoire.
- [ ] Les glossaires et les `AGENTS.md` concernés sont à jour.
- [ ] `make check` est vert.
