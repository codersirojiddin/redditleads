package reports

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/auth"
	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
	"github.com/codersirojiddin/reddit-leads/internal/opportunities"
	"github.com/codersirojiddin/reddit-leads/internal/projects"
)

type Handler struct {
	svc         *Service
	projectsSvc *projects.Service
}

func NewHandler(svc *Service, projectsSvc *projects.Service) *Handler {
	return &Handler{svc: svc, projectsSvc: projectsSvc}
}

type reportResponse struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Status      string  `json:"status"`
	Error       string  `json:"error,omitempty"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

func toReportResponse(r *Report) reportResponse {
	resp := reportResponse{
		ID:        r.ID.String(),
		ProjectID: r.ProjectID.String(),
		Status:    r.Status,
		Error:     r.Error,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.CompletedAt != nil {
		t := r.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &t
	}
	return resp
}

type opportunityResponse struct {
	ID             string  `json:"id"`
	Title          string  `json:"post_title"`
	URL            string  `json:"post_url"`
	Subreddit      string  `json:"subreddit"`
	Author         string  `json:"author"`
	RelevanceScore float64 `json:"relevance_score"`
	IntentScore    float64 `json:"intent_score"`
	TotalScore     float64 `json:"total_score"`
	Reasoning      string  `json:"reasoning"`
	SuggestedReply string  `json:"suggested_reply"`
}

func toOpportunityResponse(o *opportunities.Opportunity) opportunityResponse {
	return opportunityResponse{
		ID:             o.ID.String(),
		Title:          o.PostTitle,
		URL:            o.PostURL,
		Subreddit:      o.PostSubreddit,
		Author:         o.PostAuthor,
		RelevanceScore: o.RelevanceScore,
		IntentScore:    o.IntentScore,
		TotalScore:     o.TotalScore,
		Reasoning:      o.Reasoning,
		SuggestedReply: o.SuggestedReply,
	}
}

// Create requests a new report run for a project (POST /projects/{projectID}/reports).
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

	report, err := h.svc.RequestReport(r.Context(), projectID)
	if err != nil {
		apperr.WriteError(w, apperr.Internal("failed to request report"))
		return
	}

	apperr.JSON(w, http.StatusAccepted, toReportResponse(report))
}

// List returns all reports for a project.
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

	list, err := h.svc.ListByProject(r.Context(), projectID)
	if err != nil {
		apperr.WriteError(w, apperr.Internal("failed to list reports"))
		return
	}

	out := make([]reportResponse, 0, len(list))
	for _, rep := range list {
		out = append(out, toReportResponse(rep))
	}
	apperr.JSON(w, http.StatusOK, out)
}

type reportDetailResponse struct {
	reportResponse
	Opportunities []opportunityResponse `json:"opportunities"`
}

// Get returns a report along with its ranked opportunities
// (GET /reports/{reportID}).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	reportID, err := uuid.Parse(chi.URLParam(r, "reportID"))
	if err != nil {
		apperr.WriteError(w, apperr.BadRequest("invalid report id"))
		return
	}

	report, opps, err := h.svc.GetWithOpportunities(r.Context(), reportID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apperr.WriteError(w, apperr.NotFound("report not found"))
			return
		}
		apperr.WriteError(w, apperr.Internal("failed to load report"))
		return
	}

	// Ownership check: the report's project must belong to the caller.
	if _, err := h.projectsSvc.GetOwned(r.Context(), userID, report.ProjectID); err != nil {
		writeOwnershipErr(w, err)
		return
	}

	oppsOut := make([]opportunityResponse, 0, len(opps))
	for _, o := range opps {
		oppsOut = append(oppsOut, toOpportunityResponse(o))
	}

	apperr.JSON(w, http.StatusOK, reportDetailResponse{
		reportResponse: toReportResponse(report),
		Opportunities:  oppsOut,
	})
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
