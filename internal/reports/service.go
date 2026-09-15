package reports

import (
	"context"

	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/opportunities"
)

// TopOpportunitiesPerReport matches the product spec: rank the top 100
// opportunities into a report.
const TopOpportunitiesPerReport = 100

type Service struct {
	repo    *Repository
	oppRepo *opportunities.Repository
}

func NewService(repo *Repository, oppRepo *opportunities.Repository) *Service {
	return &Service{repo: repo, oppRepo: oppRepo}
}

// RequestReport enqueues a new report for a project (status=pending); the
// worker picks it up asynchronously.
func (s *Service) RequestReport(ctx context.Context, projectID uuid.UUID) (*Report, error) {
	return s.repo.Create(ctx, projectID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Report, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Report, error) {
	return s.repo.ListByProject(ctx, projectID)
}

// GetWithOpportunities returns the report plus its ranked opportunities (with
// post details), ready for API/email rendering.
func (s *Service) GetWithOpportunities(ctx context.Context, id uuid.UUID) (*Report, []*opportunities.Opportunity, error) {
	report, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	items, err := s.repo.ItemsByReport(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.OpportunityID)
	}

	opps, err := s.oppRepo.ByIDs(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	// Re-order opps to match report_items rank ordering.
	byID := make(map[uuid.UUID]*opportunities.Opportunity, len(opps))
	for _, o := range opps {
		byID[o.ID] = o
	}
	ordered := make([]*opportunities.Opportunity, 0, len(items))
	for _, it := range items {
		if o, ok := byID[it.OpportunityID]; ok {
			ordered = append(ordered, o)
		}
	}

	return report, ordered, nil
}
