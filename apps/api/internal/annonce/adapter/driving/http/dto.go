package http

import "time"

// The JSON contract of the annonce endpoints: keys are French.

type AnnonceResponse struct {
	ID           string               `json:"id"`
	Titre        string               `json:"titre"`
	Description  string               `json:"description"`
	Prix         int64                `json:"prix"`
	Categorie    CategorieResponse    `json:"categorie"`
	Localisation LocalisationResponse `json:"localisation"`
	Vendeur      VendeurResponse      `json:"vendeur"`
	Images       []string             `json:"images"`
	DeposeeLe    time.Time            `json:"deposeeLe"`
}

type CategorieResponse struct {
	Code    string `json:"code"`
	Libelle string `json:"libelle"`
}

type LocalisationResponse struct {
	Ville       string              `json:"ville"`
	CodePostal  string              `json:"codePostal"`
	Departement DepartementResponse `json:"departement"`
	Region      RegionResponse      `json:"region"`
}

type DepartementResponse struct {
	Code string `json:"code"`
	Nom  string `json:"nom"`
}

type RegionResponse struct {
	Code string `json:"code"`
	Nom  string `json:"nom"`
}

// RegionAvecDepartementsResponse is one region of GET /regions.
type RegionAvecDepartementsResponse struct {
	Code         string                `json:"code"`
	Nom          string                `json:"nom"`
	Departements []DepartementResponse `json:"departements"`
}

type VendeurResponse struct {
	Pseudo string `json:"pseudo"`
}

// RechercherAnnoncesResponse is one page of results of GET /annonces.
type RechercherAnnoncesResponse struct {
	Annonces []AnnonceResumeResponse `json:"annonces"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	ParPage  int                     `json:"parPage"`
}

// AnnonceResumeResponse is an annonce in a list of results.
type AnnonceResumeResponse struct {
	ID        string    `json:"id"`
	Titre     string    `json:"titre"`
	Prix      int64     `json:"prix"`
	Ville     string    `json:"ville"`
	Images    []string  `json:"images"`
	DeposeeLe time.Time `json:"deposeeLe"`
}

// DeposerAnnonceRequest is the body of POST /annonces. Prix is in centimes
// and required (a pointer tells a missing prix from a don); telephone is
// optional.
type DeposerAnnonceRequest struct {
	Titre       string `json:"titre"`
	Description string `json:"description"`
	Prix        *int64 `json:"prix"`
	Categorie   string `json:"categorie"`
	Ville       string `json:"ville"`
	CodePostal  string `json:"codePostal"`
	Pseudo      string `json:"pseudo"`
	Email       string `json:"email"`
	Telephone   string `json:"telephone"`
	MotDePasse  string `json:"motDePasse"`
}

// SupprimerAnnonceRequest is the body of POST /annonces/{id}/suppression.
// Raison is a code of the closed list.
type SupprimerAnnonceRequest struct {
	MotDePasse string `json:"motDePasse"`
	Raison     string `json:"raison"`
}

// ContacterRequest is the body of POST /annonces/{id}/contact: the acheteur's
// pseudo, email and message.
type ContacterRequest struct {
	Pseudo  string `json:"pseudo"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type DeposerAnnonceResponse struct {
	ID string `json:"id"`
}

// ErreurResponse is every error body. Champ names the request field at fault,
// when there is one, so a form can show the message under it.
type ErreurResponse struct {
	Erreur string `json:"erreur"`
	Champ  string `json:"champ,omitempty"`
}
