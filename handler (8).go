package auth

import (
	"errors"
	"net/http"

	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  userPublic `json:"user"`
}

type userPublic struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := apperr.Decode(r, &req); err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid request body"))
		return
	}

	token, user, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			apperr.WriteError(w, apperr.Conflict("email already registered"))
			return
		}
		apperr.WriteError(w, apperr.BadRequest(err.Error()))
		return
	}

	apperr.JSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  userPublic{ID: user.ID.String(), Email: user.Email},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := apperr.Decode(r, &req); err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid request body"))
		return
	}

	token, user, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		apperr.WriteError(w, apperr.Unauthorized("invalid email or password"))
		return
	}

	apperr.JSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  userPublic{ID: user.ID.String(), Email: user.Email},
	})
}
