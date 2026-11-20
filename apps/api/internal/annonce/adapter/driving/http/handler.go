package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/application/query"
	"ddd-second-hand-marketplace/internal/annonce/domain"
)

// Handler exposes the annonce use cases over HTTP.
type Handler struct {
	consulterAnnonce   query.ConsulterAnnonceQuery
	rechercherAnnonces query.RechercherAnnoncesQuery
	listerRegions      query.ListerRegionsQuery
	deposerAnnonce     command.DeposerAnnonceCommand
	supprimerAnnonce   command.SupprimerAnnonceCommand
	annonceAEditer     query.ConsulterAnnonceAEditerQuery
	corrigerAnnonce    command.CorrigerAnnonceCommand
}

func NewHandler(
	consulterAnnonce query.ConsulterAnnonceQuery,
	rechercherAnnonces query.RechercherAnnoncesQuery,
	listerRegions query.ListerRegionsQuery,
	deposerAnnonce command.DeposerAnnonceCommand,
	supprimerAnnonce command.SupprimerAnnonceCommand,
	annonceAEditer query.ConsulterAnnonceAEditerQuery,
	corrigerAnnonce command.CorrigerAnnonceCommand,
) *Handler {
	return &Handler{
		consulterAnnonce:   consulterAnnonce,
		rechercherAnnonces: rechercherAnnonces,
		listerRegions:      listerRegions,
		deposerAnnonce:     deposerAnnonce,
		supprimerAnnonce:   supprimerAnnonce,
		annonceAEditer:     annonceAEditer,
		corrigerAnnonce:    corrigerAnnonce,
	}
}

// RegisterRoutes mounts the annonce endpoints on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /annonces", h.RechercherAnnonces)
	mux.HandleFunc("GET /annonces/{id}", h.ConsulterAnnonce)
	mux.HandleFunc("POST /annonces", h.DeposerAnnonce)
	mux.HandleFunc("POST /annonces/{id}/suppression", h.SupprimerAnnonce)
	mux.HandleFunc("POST /annonces/{id}/edition", h.EditerAnnonce)
	mux.HandleFunc("POST /annonces/{id}/correction", h.CorrigerAnnonce)
	mux.HandleFunc("GET /regions", h.ListerRegions)
}

// ListerRegions handles GET /regions: every region with its departements,
// the lieux a recherche accepts.
func (h *Handler) ListerRegions(w http.ResponseWriter, r *http.Request) {
	views := h.listerRegions()
	regions := make([]RegionAvecDepartementsResponse, len(views))
	for i, view := range views {
		departements := make([]DepartementResponse, len(view.Departements))
		for j, departement := range view.Departements {
			departements[j] = DepartementResponse(departement)
		}
		regions[i] = RegionAvecDepartementsResponse{Code: view.Code, Nom: view.Nom, Departements: departements}
	}
	respond(w, http.StatusOK, regions)
}

// maxDepotBytes caps the body of a dépôt: a 4 000-character description fits
// many times over.
const maxDepotBytes = 64 << 10

// DeposerAnnonce handles POST /annonces: 201 with the new ID, or 400 naming
// the field at fault.
func (h *Handler) DeposerAnnonce(w http.ResponseWriter, r *http.Request) {
	var req DeposerAnnonceRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxDepotBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return
	}
	if req.Prix == nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Le prix est obligatoire", Champ: "prix"})
		return
	}
	id, err := h.deposerAnnonce(command.DeposerAnnonceCommandArgs{
		Titre:       req.Titre,
		Description: req.Description,
		Prix:        *req.Prix,
		Categorie:   req.Categorie,
		Ville:       req.Ville,
		CodePostal:  req.CodePostal,
		Pseudo:      req.Pseudo,
		Email:       req.Email,
		Telephone:   req.Telephone,
		MotDePasse:  req.MotDePasse,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, DeposerAnnonceResponse{ID: id})
}

// SupprimerAnnonce handles POST /annonces/{id}/suppression: 204 once the
// annonce is withdrawn, 400 naming motDePasse or raison, 404 when the annonce
// is unknown or already supprimee.
func (h *Handler) SupprimerAnnonce(w http.ResponseWriter, r *http.Request) {
	var req SupprimerAnnonceRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxDepotBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return
	}
	err := h.supprimerAnnonce(command.SupprimerAnnonceCommandArgs{
		ID:         r.PathValue("id"),
		MotDePasse: req.MotDePasse,
		Raison:     req.Raison,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EditerAnnonce handles POST /annonces/{id}/edition: the editable fields and
// the motif of a challengee annonce, for the vendeur who knows its mot de
// passe. 400 naming motDePasse, 404 when the annonce is unknown or not
// challengee.
func (h *Handler) EditerAnnonce(w http.ResponseWriter, r *http.Request) {
	var req EditerAnnonceRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxDepotBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return
	}
	view, err := h.annonceAEditer(r.PathValue("id"), req.MotDePasse)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, AnnonceAEditerResponse(*view))
}

// CorrigerAnnonce handles POST /annonces/{id}/correction: 204 once the
// annonce is back in moderation, 400 naming the field at fault, 404 when the
// annonce is unknown or not challengee.
func (h *Handler) CorrigerAnnonce(w http.ResponseWriter, r *http.Request) {
	var req CorrigerAnnonceRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxDepotBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return
	}
	if req.Prix == nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Le prix est obligatoire", Champ: "prix"})
		return
	}
	err := h.corrigerAnnonce(command.CorrigerAnnonceCommandArgs{
		ID:          r.PathValue("id"),
		MotDePasse:  req.MotDePasse,
		Titre:       req.Titre,
		Description: req.Description,
		Prix:        *req.Prix,
		Categorie:   req.Categorie,
		Ville:       req.Ville,
		CodePostal:  req.CodePostal,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ConsulterAnnonce handles GET /annonces/{id}.
func (h *Handler) ConsulterAnnonce(w http.ResponseWriter, r *http.Request) {
	view, err := h.consulterAnnonce(r.PathValue("id"))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, toAnnonceResponse(view))
}

// RechercherAnnonces handles
// GET /annonces?motCle=…&categorie=…&prixMin=…&prixMax=…&region=…&departement=…&villeOuCodePostal=…&page=…:
// one page of the activee annonces matching, most recent first. Every
// parameter is optional; the prix bounds are in centimes, region and
// departement INSEE codes.
func (h *Handler) RechercherAnnonces(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	page := 0
	if raw := params.Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			respondDomainError(w, domain.ErrPageInvalide)
			return
		}
		page = n
	}
	prixMin, err := optionalCentimes(params.Get("prixMin"), domain.ErrPrixMinInvalide)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	prixMax, err := optionalCentimes(params.Get("prixMax"), domain.ErrPrixMaxInvalide)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	view, err := h.rechercherAnnonces(query.RechercherAnnoncesArgs{
		MotCle:            params.Get("motCle"),
		Categorie:         params.Get("categorie"),
		PrixMin:           prixMin,
		PrixMax:           prixMax,
		Region:            params.Get("region"),
		Departement:       params.Get("departement"),
		VilleOuCodePostal: params.Get("villeOuCodePostal"),
		Page:              page,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	annonces := make([]AnnonceResumeResponse, len(view.Annonces))
	for i, a := range view.Annonces {
		annonces[i] = AnnonceResumeResponse(a)
	}
	respond(w, http.StatusOK, RechercherAnnoncesResponse{
		Annonces: annonces,
		Total:    view.Total,
		Page:     view.Page,
		ParPage:  view.ParPage,
	})
}

// optionalCentimes reads an optional query parameter as an integer number of
// centimes: nil when absent, errInvalide when it is no integer.
func optionalCentimes(raw string, errInvalide error) (*int64, error) {
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errInvalide
	}
	return &n, nil
}

func toAnnonceResponse(view *query.AnnonceView) AnnonceResponse {
	return AnnonceResponse{
		ID:          view.ID,
		Titre:       view.Titre,
		Description: view.Description,
		Prix:        view.Prix,
		Categorie:   CategorieResponse{Code: view.Categorie.Code, Libelle: view.Categorie.Libelle},
		Localisation: LocalisationResponse{
			Ville:       view.Localisation.Ville,
			CodePostal:  view.Localisation.CodePostal,
			Departement: DepartementResponse(view.Localisation.Departement),
			Region:      RegionResponse(view.Localisation.Region),
		},
		Vendeur:   VendeurResponse{Pseudo: view.Vendeur.Pseudo},
		Images:    view.Images,
		DeposeeLe: view.DeposeeLe,
	}
}

// invalidFields maps each validation error of the domain to the JSON key of
// the field at fault and the message a form shows under it.
var invalidFields = []struct {
	err     error
	champ   string
	message string
}{
	{domain.ErrTitreInvalide, "titre", "Le titre doit comporter de 3 à 100 caractères"},
	{domain.ErrDescriptionInvalide, "description", "La description doit comporter de 10 à 4 000 caractères"},
	{domain.ErrPrixNegatif, "prix", "Le prix ne peut pas être négatif"},
	{domain.ErrCategorieInconnue, "categorie", "Catégorie inconnue"},
	{domain.ErrVilleVide, "ville", "La ville est obligatoire"},
	{domain.ErrCodePostalInvalide, "codePostal", "Le code postal doit comporter 5 chiffres"},
	{domain.ErrDepartementInconnu, "codePostal", "Le code postal ne désigne aucun département"},
	{domain.ErrPseudoInvalide, "pseudo", "Le pseudo doit comporter de 2 à 30 caractères"},
	{domain.ErrEmailInvalide, "email", "L'email n'est pas valide"},
	{domain.ErrPageInvalide, "page", "Le numéro de page doit être un entier positif"},
	{domain.ErrPrixMinInvalide, "prixMin", "Le prix minimum doit être un montant positif ou nul"},
	{domain.ErrPrixMaxInvalide, "prixMax", "Le prix maximum doit être un montant positif ou nul"},
	{domain.ErrFourchettePrixInvalide, "prixMin", "Le prix minimum ne peut pas dépasser le prix maximum"},
	{domain.ErrRegionInconnue, "region", "Région inconnue"},
	{domain.ErrCodeDepartementInconnu, "departement", "Département inconnu"},
	{domain.ErrLieuAmbigu, "region", "Choisissez une région ou un département, pas les deux"},
	{domain.ErrMotDePasseTropCourt, "motDePasse", "Le mot de passe de l'annonce doit comporter au moins 8 caractères"},
	{domain.ErrMotDePasseIncorrect, "motDePasse", "Le mot de passe de l'annonce est incorrect"},
	{domain.ErrRaisonSuppressionInvalide, "raison", "Choisissez une raison de suppression"},
	{domain.ErrTelephoneInvalide, "telephone", "Le téléphone doit comporter 10 chiffres et commencer par 0"},
}

// respondDomainError maps a domain error to its status code and a French
// message: a validation error is a 400 naming its field. Anything else is a
// 500 whose detail stays in the logs.
func respondDomainError(w http.ResponseWriter, err error) {
	for _, f := range invalidFields {
		if errors.Is(err, f.err) {
			respond(w, http.StatusBadRequest, ErreurResponse{Erreur: f.message, Champ: f.champ})
			return
		}
	}
	switch {
	case errors.Is(err, domain.ErrAnnonceIntrouvable):
		respond(w, http.StatusNotFound, ErreurResponse{Erreur: "Annonce introuvable"})
	case errors.Is(err, domain.ErrIDAnnonceInvalide):
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Identifiant d'annonce invalide"})
	default:
		log.Printf("annonce: unexpected error: %v", err)
		respond(w, http.StatusInternalServerError, ErreurResponse{Erreur: "Erreur interne"})
	}
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("annonce: encoding the response failed: %v", err)
	}
}
