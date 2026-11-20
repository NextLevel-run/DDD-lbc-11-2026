package http

import "time"

// The JSON contract of the moderation endpoints: keys are French.

type ErreurResponse struct {
	Erreur string `json:"erreur"`
	Champ  string `json:"champ,omitempty"`
}

// ModerateurRequest is the body of the lock and release endpoints.
type ModerateurRequest struct {
	Moderateur string `json:"moderateur"`
}

// ChallengeRequest is the body of the challenge endpoint.
type ChallengeRequest struct {
	Moderateur string `json:"moderateur"`
	Motif      string `json:"motif"`
}

type AnnonceAModererResumeResponse struct {
	ID               string     `json:"id"`
	Titre            string     `json:"titre"`
	Prix             int64      `json:"prix"`
	Ville            string     `json:"ville"`
	Pseudo           string     `json:"pseudo"`
	MisEnFileLe      time.Time  `json:"misEnFileLe"`
	VerrouillePar    string     `json:"verrouillePar"`
	VerrouilleDepuis *time.Time `json:"verrouilleDepuis,omitempty"`
}

type ListerAnnoncesAModererResponse struct {
	Annonces []AnnonceAModererResumeResponse `json:"annonces"`
}

type EvenementResponse struct {
	Type       string    `json:"type"`
	Date       time.Time `json:"date"`
	Moderateur string    `json:"moderateur"`
	Motif      string    `json:"motif"`
}

type AnnonceAModererResponse struct {
	ID               string              `json:"id"`
	Titre            string              `json:"titre"`
	Description      string              `json:"description"`
	Prix             int64               `json:"prix"`
	Categorie        string              `json:"categorie"`
	Ville            string              `json:"ville"`
	CodePostal       string              `json:"codePostal"`
	Pseudo           string              `json:"pseudo"`
	DeposeeLe        time.Time           `json:"deposeeLe"`
	Statut           string              `json:"statut"`
	VerrouillePar    string              `json:"verrouillePar"`
	VerrouilleDepuis *time.Time          `json:"verrouilleDepuis,omitempty"`
	Historique       []EvenementResponse `json:"historique"`
}
