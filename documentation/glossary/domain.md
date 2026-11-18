# Glossaire du domaine

Langage ubiquitaire de la place de marché d'occasion, en français : c'est la langue dans laquelle
l'équipe modélise, et le code reprend ces mots sans les traduire. Un seul nom par concept.
Tags (en anglais, ce sont les termes de `technical.md`) : _aggregate_, _value object_, _command_,
_event_, _rule_, _actor_.

Une entrée dit ce qu'est un terme, en une ou deux phrases, 300 caractères au plus. Un terme mérite
une entrée quand il nomme un agrégat, un sous-domaine ou bounded context, ou un concept clé partagé
par plusieurs modules ou montré aux utilisateurs. Un réglage, une option, une variante d'une autre
entrée ou un détail de règle n'en a pas : sa place est dans l'`AGENTS.md` de son module ou dans le
code. Une entrée se suffit à elle-même : aucun renvoi vers une spec ou un fichier.

Une entrée s'écrit ainsi, sous un titre par sous-domaine :

    - **Annonce déposée** (`AnnonceDeposee`) — _event_. Définition.

Le terme est écrit comme on le dit, avec ses accents ; la parenthèse donne son nom dans le code
quand il diffère (sans accent ni espace). Ajouter un terme dès qu'il est tranché, dans le même
changement que le code qui l'utilise ; ne jamais inventer dans le code un nom absent d'ici, ni le
traduire en anglais.

## Carte du domaine

| Sous-domaine | Type | Bounded context | Agrégats |
| ------------ | ---- | --------------- | -------- |
