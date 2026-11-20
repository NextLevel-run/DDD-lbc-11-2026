// Package http exposes the moderation use cases. The routes have no
// authentication: the page that calls them is hidden, not protected.
package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/application/query"
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// Handler exposes the moderation use cases over HTTP.
type Handler struct {
	listerAnnonces   query.ListerAnnoncesAModererQuery
	consulterAnnonce query.ConsulterAnnonceAModererQuery
	verrouiller      command.VerrouillerAnnonceCommand
	relacher         command.RelacherVerrouCommand
	valider          command.ValiderAnnonceCommand
	refuser          command.RefuserAnnonceCommand
	challenger       command.ChallengerAnnonceCommand
}

func NewHandler(
	listerAnnonces query.ListerAnnoncesAModererQuery,
	consulterAnnonce query.ConsulterAnnonceAModererQuery,
	verrouiller command.VerrouillerAnnonceCommand,
	relacher command.RelacherVerrouCommand,
	valider command.ValiderAnnonceCommand,
	refuser command.RefuserAnnonceCommand,
	challenger command.ChallengerAnnonceCommand,
) *Handler {
	return &Handler{
		listerAnnonces:   listerAnnonces,
		consulterAnnonce: consulterAnnonce,
		verrouiller:      verrouiller,
		relacher:         relacher,
		valider:          valider,
		refuser:          refuser,
		challenger:       challenger,
	}
}

// RegisterRoutes mounts the moderation endpoints on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /moderation/annonces", h.ListerAnnonces)
	mux.HandleFunc("GET /moderation/annonces/{id}", h.ConsulterAnnonce)
	mux.HandleFunc("POST /moderation/annonces/{id}/verrou", h.Verrouiller)
	mux.HandleFunc("POST /moderation/annonces/{id}/relachement", h.Relacher)
	mux.HandleFunc("POST /moderation/annonces/{id}/validation", h.Valider)
	mux.HandleFunc("POST /moderation/annonces/{id}/refus", h.Refuser)
	mux.HandleFunc("POST /moderation/annonces/{id}/challenge", h.Challenger)
}

// ListerAnnonces handles GET /moderation/annonces: the queue, oldest first.
func (h *Handler) ListerAnnonces(w http.ResponseWriter, r *http.Request) {
	views, err := h.listerAnnonces()
	if err != nil {
		respondDomainError(w, err)
		return
	}
	annonces := make([]AnnonceAModererResumeResponse, len(views))
	for i, v := range views {
		annonces[i] = AnnonceAModererResumeResponse{
			ID: v.ID, Titre: v.Titre, Prix: v.Prix, Ville: v.Ville, Pseudo: v.Pseudo,
			MisEnFileLe: v.MisEnFileLe, VerrouillePar: v.VerrouillePar, VerrouilleDepuis: optionalTime(v.VerrouilleDepuis),
		}
	}
	respond(w, http.StatusOK, ListerAnnoncesAModererResponse{Annonces: annonces})
}

// ConsulterAnnonce handles GET /moderation/annonces/{id}: 404 when unknown.
func (h *Handler) ConsulterAnnonce(w http.ResponseWriter, r *http.Request) {
	v, err := h.consulterAnnonce(r.PathValue("id"))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	historique := make([]EvenementResponse, len(v.Historique))
	for i, e := range v.Historique {
		historique[i] = EvenementResponse(e)
	}
	respond(w, http.StatusOK, AnnonceAModererResponse{
		ID: v.ID, Titre: v.Titre, Description: v.Description, Prix: v.Prix, Categorie: v.Categorie,
		Ville: v.Ville, CodePostal: v.CodePostal, Pseudo: v.Pseudo, DeposeeLe: v.DeposeeLe, Statut: v.Statut,
		VerrouillePar: v.VerrouillePar, VerrouilleDepuis: optionalTime(v.VerrouilleDepuis), Historique: historique,
	})
}

// Verrouiller handles POST /moderation/annonces/{id}/verrou {moderateur}: 204,
// 400 on moderateur, 404 or 409.
func (h *Handler) Verrouiller(w http.ResponseWriter, r *http.Request) {
	moderateur, ok := readModerateur(w, r)
	if !ok {
		return
	}
	if err := h.verrouiller(command.VerrouillerAnnonceCommandArgs{AnnonceID: r.PathValue("id"), Moderateur: moderateur}); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Relacher handles POST /moderation/annonces/{id}/relachement {moderateur}:
// 204, 400 on moderateur, 404 or 409.
func (h *Handler) Relacher(w http.ResponseWriter, r *http.Request) {
	moderateur, ok := readModerateur(w, r)
	if !ok {
		return
	}
	if err := h.relacher(command.RelacherVerrouCommandArgs{AnnonceID: r.PathValue("id"), Moderateur: moderateur}); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Valider handles POST /moderation/annonces/{id}/validation {moderateur}: 204,
// 400 on moderateur, 404 or 409.
func (h *Handler) Valider(w http.ResponseWriter, r *http.Request) {
	moderateur, ok := readModerateur(w, r)
	if !ok {
		return
	}
	if err := h.valider(command.ValiderAnnonceCommandArgs{AnnonceID: r.PathValue("id"), Moderateur: moderateur}); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Refuser handles POST /moderation/annonces/{id}/refus {moderateur}: 204, 400
// on moderateur, 404 or 409.
func (h *Handler) Refuser(w http.ResponseWriter, r *http.Request) {
	moderateur, ok := readModerateur(w, r)
	if !ok {
		return
	}
	if err := h.refuser(command.RefuserAnnonceCommandArgs{AnnonceID: r.PathValue("id"), Moderateur: moderateur}); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Challenger handles POST /moderation/annonces/{id}/challenge {moderateur,
// motif}: 204, 400 on moderateur or motif, 404 or 409.
func (h *Handler) Challenger(w http.ResponseWriter, r *http.Request) {
	var req ChallengeRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return
	}
	err := h.challenger(command.ChallengerAnnonceCommandArgs{AnnonceID: r.PathValue("id"), Moderateur: req.Moderateur, Motif: req.Motif})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// maxBodyBytes caps a request body: a pseudo fits many times over, a motif too (1 000 characters).
const maxBodyBytes = 4 << 10

func readModerateur(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req ModerateurRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Requête invalide"})
		return "", false
	}
	return req.Moderateur, true
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// conflicts are the errors of lock and state the API answers with a 409.
var conflicts = []struct {
	err     error
	message string
}{
	{domain.ErrAnnonceDejaVerrouillee, "Cette annonce est déjà verrouillée par un autre modérateur"},
	{domain.ErrAnnoncePasEnFile, "Cette annonce n'est pas dans la file de modération"},
	{domain.ErrAnnoncePasVerrouilleeParModerateur, "Cette annonce n'est pas verrouillée par vous"},
	{domain.ErrVerrouExpire, "Votre verrou a expiré"},
	{domain.ErrConflitVersion, "L'annonce vient d'être modifiée, réessayez"},
}

// respondDomainError maps a domain error to its status and a French message.
// Anything unknown is a 500 whose detail stays in the logs.
func respondDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrModerateurInvalide):
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Le pseudo du modérateur doit comporter de 2 à 30 caractères", Champ: "moderateur"})
		return
	case errors.Is(err, domain.ErrMotifInvalide):
		respond(w, http.StatusBadRequest, ErreurResponse{Erreur: "Le motif doit comporter de 10 à 1 000 caractères", Champ: "motif"})
		return
	case errors.Is(err, domain.ErrAnnonceIntrouvable):
		respond(w, http.StatusNotFound, ErreurResponse{Erreur: "Annonce introuvable"})
		return
	}
	for _, c := range conflicts {
		if errors.Is(err, c.err) {
			respond(w, http.StatusConflict, ErreurResponse{Erreur: c.message})
			return
		}
	}
	log.Printf("moderation: unexpected error: %v", err)
	respond(w, http.StatusInternalServerError, ErreurResponse{Erreur: "Erreur interne"})
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("moderation: encoding the response failed: %v", err)
	}
}
