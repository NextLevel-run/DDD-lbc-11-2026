package domain

import "time"

// DureeVerrou is how long a lock holds: past it, the annonce is free again.
const DureeVerrou = 2 * time.Hour

// Etat of an annonce in the moderation flow, derived by replaying its stream.
type Etat string

const (
	EtatVide        Etat = ""            // no event yet: unknown here
	EtatEnFile      Etat = "en-file"     // waiting for a moderator
	EtatVerrouillee Etat = "verrouillee" // a moderator holds it
	EtatValidee     Etat = "validee"     // decided: online
	EtatRefusee     Etat = "refusee"     // decided: rejected for fraud
	EtatChallengee  Etat = "challengee"  // decided: back to the vendeur
)

// HistoriqueAnnonce is the event-sourced aggregate: the stream of one annonce.
// A decision method checks the state the replay gives, then appends an event;
// nothing mutates a stored state. The version is the number of events already
// stored, what the repository checks on append.
type HistoriqueAnnonce struct {
	annonceID  string
	evenements []Evenement
	version    int

	etat        Etat
	annonce     InstantaneAnnonce
	misEnFileLe time.Time
	moderateur  string
	depuis      time.Time
}

// NewHistoriqueAnnonce starts an empty stream.
func NewHistoriqueAnnonce(annonceID string) *HistoriqueAnnonce {
	return &HistoriqueAnnonce{annonceID: annonceID}
}

// RestoreHistoriqueAnnonce rebuilds the aggregate by replaying the stored
// events, oldest first.
func RestoreHistoriqueAnnonce(annonceID string, evenements []Evenement) *HistoriqueAnnonce {
	h := NewHistoriqueAnnonce(annonceID)
	for _, e := range evenements {
		h.appliquer(e)
	}
	h.version = len(evenements)
	return h
}

func (h *HistoriqueAnnonce) appliquer(e Evenement) {
	h.evenements = append(h.evenements, e)
	switch e := e.(type) {
	case AnnonceMiseEnFileEvent:
		h.etat, h.annonce, h.misEnFileLe = EtatEnFile, e.Annonce, e.Date
		h.moderateur, h.depuis = "", time.Time{}
	case AnnonceVerrouilleeEvent:
		h.etat, h.moderateur, h.depuis = EtatVerrouillee, e.Moderateur, e.Date
	case VerrouAnnonceRelacheEvent:
		h.etat, h.moderateur, h.depuis = EtatEnFile, "", time.Time{}
	case AnnonceValideeEvent:
		h.etat, h.moderateur, h.depuis = EtatValidee, "", time.Time{}
	case AnnonceRefuseeEvent:
		h.etat, h.moderateur, h.depuis = EtatRefusee, "", time.Time{}
	case AnnonceChallengeeEvent:
		h.etat, h.moderateur, h.depuis = EtatChallengee, "", time.Time{}
	}
}

// decider records a new event: applied to the state and kept to be appended.
func (h *HistoriqueAnnonce) decider(e Evenement) { h.appliquer(e) }

func (h *HistoriqueAnnonce) AnnonceID() string          { return h.annonceID }
func (h *HistoriqueAnnonce) Etat() Etat                 { return h.etat }
func (h *HistoriqueAnnonce) Annonce() InstantaneAnnonce { return h.annonce }
func (h *HistoriqueAnnonce) MisEnFileLe() time.Time     { return h.misEnFileLe }
func (h *HistoriqueAnnonce) Version() int               { return h.version }

// Evenements returns the whole stream, stored then new, oldest first.
func (h *HistoriqueAnnonce) Evenements() []Evenement {
	return append([]Evenement(nil), h.evenements...)
}

// NouveauxEvenements returns the events decided since the aggregate was
// restored: what the repository has to append.
func (h *HistoriqueAnnonce) NouveauxEvenements() []Evenement {
	return append([]Evenement(nil), h.evenements[h.version:]...)
}

// Verrou returns the holder and the date of a lock still valid at now.
func (h *HistoriqueAnnonce) Verrou(now time.Time) (moderateur string, depuis time.Time, ok bool) {
	if h.etat != EtatVerrouillee || h.verrouExpire(now) {
		return "", time.Time{}, false
	}
	return h.moderateur, h.depuis, true
}

// EstAModerer tells whether the annonce is in the queue, locked or not.
func (h *HistoriqueAnnonce) EstAModerer() bool {
	return h.etat == EtatEnFile || h.etat == EtatVerrouillee
}

func (h *HistoriqueAnnonce) verrouExpire(now time.Time) bool {
	return now.Sub(h.depuis) > DureeVerrou
}

// MettreEnFile puts a deposited annonce, or a challenged one the vendeur has
// corrected, in the queue; the snapshot replaces the previous one. Any other
// state (already queued, locked, validee, refusee) is left alone: a
// redelivered or late event is not an error.
func (h *HistoriqueAnnonce) MettreEnFile(annonce InstantaneAnnonce, now time.Time) {
	if h.etat != EtatVide && h.etat != EtatChallengee {
		return
	}
	h.decider(AnnonceMiseEnFileEvent{Annonce: annonce, Date: now})
}

// Verrouiller gives the annonce to a moderator. A lock older than 2 hours is
// released first (cause expiration). A valid lock of another moderator is
// ErrAnnonceDejaVerrouillee; the holder asking again changes nothing.
func (h *HistoriqueAnnonce) Verrouiller(moderateur Moderateur, now time.Time) error {
	switch h.etat {
	case EtatVide:
		return ErrAnnonceIntrouvable
	case EtatVerrouillee:
		if h.verrouExpire(now) {
			h.decider(VerrouAnnonceRelacheEvent{Cause: CauseExpiration, Date: now})
			break
		}
		if h.moderateur == moderateur.Pseudo() {
			return nil
		}
		return ErrAnnonceDejaVerrouillee
	case EtatEnFile:
	default:
		return ErrAnnoncePasEnFile
	}
	h.decider(AnnonceVerrouilleeEvent{Moderateur: moderateur.Pseudo(), Date: now})
	return nil
}

// Relacher puts the annonce back in the queue. Only the holder of a valid
// lock may.
func (h *HistoriqueAnnonce) Relacher(moderateur Moderateur, now time.Time) error {
	if err := h.exigerVerrou(moderateur, now); err != nil {
		return err
	}
	h.decider(VerrouAnnonceRelacheEvent{Moderateur: moderateur.Pseudo(), Cause: CauseManuel, Date: now})
	return nil
}

// exigerVerrou checks that moderateur holds a valid lock.
func (h *HistoriqueAnnonce) exigerVerrou(moderateur Moderateur, now time.Time) error {
	if h.etat == EtatVide {
		return ErrAnnonceIntrouvable
	}
	if h.etat != EtatVerrouillee || h.moderateur != moderateur.Pseudo() {
		return ErrAnnoncePasVerrouilleeParModerateur
	}
	if h.verrouExpire(now) {
		return ErrVerrouExpire
	}
	return nil
}

// Valider puts the annonce online: only the holder of a valid lock may. It
// returns the event it appended, for the command to publish once stored.
func (h *HistoriqueAnnonce) Valider(moderateur Moderateur, now time.Time) (AnnonceValideeEvent, error) {
	if err := h.exigerVerrou(moderateur, now); err != nil {
		return AnnonceValideeEvent{}, err
	}
	e := NewAnnonceValideeEvent(h.annonceID, moderateur.Pseudo(), now)
	h.decider(e)
	return e, nil
}

// Refuser rejects the annonce for good (fraud): only the holder of a valid
// lock may.
func (h *HistoriqueAnnonce) Refuser(moderateur Moderateur, now time.Time) (AnnonceRefuseeEvent, error) {
	if err := h.exigerVerrou(moderateur, now); err != nil {
		return AnnonceRefuseeEvent{}, err
	}
	e := NewAnnonceRefuseeEvent(h.annonceID, moderateur.Pseudo(), now)
	h.decider(e)
	return e, nil
}

// Challenger sends the annonce back to its vendeur with a motif: only the
// holder of a valid lock may.
func (h *HistoriqueAnnonce) Challenger(moderateur Moderateur, motif Motif, now time.Time) (AnnonceChallengeeEvent, error) {
	if err := h.exigerVerrou(moderateur, now); err != nil {
		return AnnonceChallengeeEvent{}, err
	}
	e := NewAnnonceChallengeeEvent(h.annonceID, moderateur.Pseudo(), motif.Texte(), now)
	h.decider(e)
	return e, nil
}
