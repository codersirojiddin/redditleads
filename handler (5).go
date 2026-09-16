package users

import (
	"net/http"

	"github.com/codersirojiddin/reddit-leads/internal/auth"
	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type profileResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apperr.WriteError(w, apperr.Unauthorized("not authenticated"))
		return
	}

	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		apperr.WriteError(w, apperr.NotFound("user not found"))
		return
	}

	apperr.JSON(w, http.StatusOK, profileResponse{ID: profile.ID.String(), Email: profile.Email})
}
