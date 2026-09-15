package keywords

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/auth"
	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
	"github.com/codersirojiddin/reddit-leads/internal/projects"
)

type Handler struct {
	svc         *Service
	projectsSvc *projects.Service
}

func NewHandler(svc *Service, projectsSvc *projects.Service) *Handler {
	return &Handler{svc: svc, projectsSvc: projectsSvc}
}

type createRequest struct {
	Keyword string `json:"keyword"`
}

type keywordResponse struct {
	ID      string `json:"id"`
	Keyword string `json:"keyword"`
}

func toResponse(k *Keyword) keywordResponse {
	return keywordResponse{ID: k.ID.String(), Keyword: k.Keyword}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid project id"))
		return
	}
	if _, err := h.projectsSvc.GetOwned(r.Context(), userID, projectID); err != nil {
		writeOwnershipErr(w, err)
		return
	}

	var req createRequest
	if err := apperr.Decode(r, &req); err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid request body"))
		return
	}

	k, err := h.svc.Add(r.Context(), projectID, req.Keyword)
	if err != nil {
		switch {
		case errors.Is(err, ErrLimitReached), errors.Is(err, ErrEmpty):
			apperr.WriteError(w, apperr.BadRequest(err.Error()))
		case errors.Is(err, ErrDuplicate):
			apperr.WriteError(w, apperr.Conflict(err.Error()))
		default:
			apperr.WriteError(w, apperr.Internal("failed to add keyword"))
		}
		return
	}

	apperr.JSON(w, http.StatusCreated, toResponse(k))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid project id"))
		return
	}
	if _, err := h.projectsSvc.GetOwned(r.Context(), userID, projectID); err != nil {
		writeOwnershipErr(w, err)
		return
	}

	list, err := h.svc.List(r.Context(), projectID)
	if err != nil {
		apperr.WriteError(w, apperr.Internal("failed to list keywords"))
		return
	}

	out := make([]keywordResponse, 0, len(list))
	for _, k := range list {
		out = append(out, toResponse(k))
	}
	apperr.JSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid project id"))
		return
	}
	if _, err := h.projectsSvc.GetOwned(r.Context(), userID, projectID); err != nil {
		writeOwnershipErr(w, err)
		return
	}

	keywordID, err := uuid.Parse(chi.URLParam(r, "keywordID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid keyword id"))
		return
	}

	if err := h.svc.Delete(r.Context(), keywordID); err != nil {
		apperr.WriteError(w, apperr.Internal("failed to delete keyword"))
		return
	}
	apperr.JSON(w, http.StatusNoContent, nil)
}

func writeOwnershipErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, projects.ErrNotFound):
		apperr.WriteError(w, apperr.NotFound("project not found"))
	case errors.Is(err, projects.ErrForbidden):
		apperr.WriteError(w, apperr.Forbidden("you do not have access to this project"))
	default:
		apperr.WriteError(w, apperr.Internal("something went wrong"))
	}
}
