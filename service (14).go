package keywords

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// MaxKeywordsPerProject matches the product spec: 5 keywords per project.
const MaxKeywordsPerProject = 5

var ErrLimitReached = errors.New("a project can have at most 5 keywords")
var ErrEmpty = errors.New("keyword cannot be empty")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, projectID uuid.UUID, keyword string) (*Keyword, error) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return nil, ErrEmpty
	}

	count, err := s.repo.CountByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if count >= MaxKeywordsPerProject {
		return nil, ErrLimitReached
	}

	return s.repo.Create(ctx, projectID, keyword)
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID) ([]*Keyword, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
