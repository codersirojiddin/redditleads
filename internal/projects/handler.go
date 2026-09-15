package projects

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/auth"
	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Name               string `json:"name"`
	ProductDescription string `json:"product_description"`
	TargetURL          string `json:"target_url"`
}

type projectResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ProductDescription string `json:"product_description"`
	TargetURL          string `json:"target_url"`
	CreatedAt          string `json:"created_at"`
}

func toResponse(p *Project) projectResponse {
	return projectResponse{
		ID:                 p.ID.String(),
		Name:               p.Name,
		ProductDescription: p.ProductDescription,
		TargetURL:          p.TargetURL,
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	var req createRequest
	if err := apperr.Decode(r, &req); err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid request body"))
		return
	}

	p, err := h.svc.Create(r.Context(), userID, req.Name, req.ProductDescription, req.TargetURL)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			apperr.WriteError(w, apperr.BadRequest(err.Error()))
			return
		}
		apperr.WriteError(w, apperr.Internal("failed to create project"))
		return
	}

	apperr.JSON(w, http.StatusCreated, toResponse(p))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	list, err := h.svc.List(r.Context(), userID)
	if err != nil {
		apperr.WriteError(w, apperr.Internal("failed to list projects"))
		return
	}

	out := make([]projectResponse, 0, len(list))
	for _, p := range list {
		out = append(out, toResponse(p))
	}
	apperr.JSON(w, http.StatusOK, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid project id"))
		return
	}

	p, err := h.svc.GetOwned(r.Context(), userID, projectID)
	if err != nil {
		writeProjectErr(w, err)
		return
	}
	apperr.JSON(w, http.StatusOK, toResponse(p))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid project id"))
		return
	}

	if err := h.svc.Delete(r.Context(), userID, projectID); err != nil {
		writeProjectErr(w, err)
		return
	}
	apperr.JSON(w, http.StatusNoContent, nil)
}

func writeProjectErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		apperr.WriteError(w, apperr.NotFound("project not found"))
	case errors.Is(err, ErrForbidden):
		apperr.WriteError(w, apperr.Forbidden("you do not have access to this project"))
	default:
		apperr.WriteError(w, apperr.Internal("something went wrong"))
	}
}
