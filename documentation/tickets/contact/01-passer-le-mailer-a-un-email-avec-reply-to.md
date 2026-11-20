# 01: Passer le port Mailer à un email avec Reply-To

**What to build:** un préfactoring qui prépare le contact du vendeur. Le port `Mailer` reçoit désormais un seul `Email` qui porte le destinataire, l'expéditeur, une adresse de réponse facultative, l'objet et le corps. Un mail envoyé avec une adresse de réponse permet à son destinataire de cliquer sur « Répondre » pour écrire à cette adresse, et non à l'expéditeur fixe de la plateforme. Rien ne change pour le vendeur : l'email de confirmation de dépôt part comme avant.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] Le port `Mailer` expose une seule façon d'envoyer un mail simple, qui prend une struct `Email{To, From, ReplyTo, Title, Body}`. `ReplyTo` est facultatif.
- [x] L'adapter SMTP écrit l'en-tête `Reply-To` quand `ReplyTo` est renseigné, et seulement dans ce cas. Un test le vérifie dans les deux cas.
- [x] Le fake mailer logge le `ReplyTo`, et le spy l'enregistre pour que les tests puissent l'affirmer.
- [x] L'envoi de la confirmation de dépôt passe par le nouveau port, sans `ReplyTo`. Son objet, son corps, son expéditeur et ses liens ne changent pas, et ses tests passent sans retouche de leurs attentes.
- [x] Il ne reste aucun appel à l'ancienne signature.
- [x] La ligne **Mailer** du glossaire technique est à jour si sa définition change.
- [x] `make check` est vert.
