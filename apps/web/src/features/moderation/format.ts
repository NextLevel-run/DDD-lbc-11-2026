const dateHeure = new Intl.DateTimeFormat('fr-FR', {
  dateStyle: 'long',
  timeStyle: 'short',
  timeZone: 'Europe/Paris',
})

// A date and its time, in the Paris time zone so server and browser render
// the same text.
export function formatDateHeure(iso: string): string {
  return dateHeure.format(new Date(iso))
}

// What the history says of an event of the stream.
export function libelleEvenement(
  type: string,
  moderateur: string,
  motif = '',
): string {
  switch (type) {
    case 'AnnonceMiseEnFile':
      return 'Mise en file de modération'
    case 'AnnonceVerrouillee':
      return `Verrouillée par ${moderateur}`
    case 'VerrouAnnonceRelache':
      return moderateur
        ? `Verrou relâché par ${moderateur}`
        : 'Verrou expiré après 2 heures'
    case 'AnnonceValidee':
      return `Validée par ${moderateur}`
    case 'AnnonceRefusee':
      return `Refusée par ${moderateur}`
    case 'AnnonceChallengee':
      return `Challengée par ${moderateur} : ${motif}`
    default:
      return type
  }
}
