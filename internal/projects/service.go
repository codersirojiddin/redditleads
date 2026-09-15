package projects

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("you do not have access to this project")
var ErrValidation = errors.New("name and product description are required")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, name, description, targetURL string) (*Project, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || description == "" {
		return nil, ErrValidation
	}

	return s.repo.Create(ctx, &Project{
		UserID:             userID,
		Name:               name,
		ProductDescription: description,
		TargetURL:          strings.TrimSpace(targetURL),
	})
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]*Project, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetOwned fetches a project and verifies it belongs to userID.
func (s *Service) GetOwned(ctx context.Context, userID, projectID uuid.UUID) (*Project, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrForbidden
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, userID, projectID uuid.UUID) error {
	if _, err := s.GetOwned(ctx, userID, projectID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, projectID)
}
