# 03: Valider, refuser ou challenger une annonce

**What to build:** le modérateur qui détient le verrou décide du sort de l'annonce. Valider la met en ligne et prévient le vendeur par e-mail. La refuser la laisse invisible, sans e-mail (c'est de la fraude). La challenger, avec un motif, prévient le vendeur par e-mail en lui indiquant le motif et le lien de correction.

**Blocked by:** 02 (Voir et verrouiller les annonces à modérer)

**Status:** ready-for-agent

- [ ] Valider, refuser et challenger exigent un verrou valide détenu par le modérateur : sinon `ErrAnnoncePasVerrouilleeParModerateur` ou `ErrVerrouExpire` (409). Une fois la décision prise, l'annonce n'est plus dans la file.
- [ ] Challenger exige un motif de 10 à 1 000 caractères une fois rogné (`ErrMotifInvalide`, 400 sur `motif`).
- [ ] Chaque commande enregistre l'événement dans le flux, puis publie `AnnonceValidee`, `AnnonceRefusee` ou `AnnonceChallengee` (accesseurs `AnnonceID()` et, pour la dernière, `Motif()`). Les autres événements restent dans le flux.
- [ ] L'API expose `POST /moderation/annonces/{id}/validation`, `…/refus` et `…/challenge`.
- [ ] Le contexte `annonce` consomme ces événements par des interfaces locales. `AnnonceValidee` active l'annonce, qui devient visible, et envoie « Votre annonce « {titre} » est validée » avec le lien vers sa page. `AnnonceChallengee` passe l'annonce à `Challengee`, conserve le motif et envoie un e-mail avec le motif et le lien `/annonces/{id}/corriger`. `AnnonceRefusee` passe l'annonce à `Refusee` et n'envoie rien.
- [ ] Un échec d'envoi est journalisé, jamais réessayé, et ne fait pas échouer la décision.
- [ ] La page `/moderation/$id` propose Valider, Refuser et Challenger (motif obligatoire) selon l'état et le verrou.
- [ ] Les tests couvrent les trois décisions de bout en bout sur `server_test.go` (visibilité avant et après, e-mails envoyés ou non), les erreurs de verrou, le motif invalide, et les pages sur la gateway en mémoire.
- [ ] Les glossaires et les `AGENTS.md` concernés sont à jour.
- [ ] `make check` est vert.
