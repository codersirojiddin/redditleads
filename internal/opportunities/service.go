package opportunities

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/ai"
	"github.com/codersirojiddin/reddit-leads/internal/discovery"
	"github.com/codersirojiddin/reddit-leads/internal/scoring"
)

type Service struct {
	repo       *Repository
	provider   discovery.Provider
	classifier *ai.Classifier
	// MaxPostsPerKeyword caps how many posts from a single keyword search
	// get scored, to bound LLM spend per run.
	MaxPostsPerKeyword int
	// KeywordSearchDelay is waited between each keyword's Reddit search to
	// stay under Reddit's public-endpoint rate limit (no OAuth app, so this
	// is more conservative than the authenticated API would need).
	KeywordSearchDelay time.Duration
}

func NewService(repo *Repository, provider discovery.Provider, classifier *ai.Classifier, maxPostsPerKeyword int) *Service {
	if maxPostsPerKeyword <= 0 {
		maxPostsPerKeyword = 50
	}
	return &Service{
		repo:               repo,
		provider:           provider,
		classifier:         classifier,
		MaxPostsPerKeyword: maxPostsPerKeyword,
		KeywordSearchDelay: 6 * time.Second,
	}
}

// KeywordInput is a (id, text) pair for a project's keywords.
type KeywordInput struct {
	ID   uuid.UUID
	Text string
}

// Run discovers Reddit posts for each keyword, scores each new post, and
// persists the results as opportunities for projectID. It returns how many
// new opportunities were created.
func (s *Service) Run(ctx context.Context, projectID uuid.UUID, productDescription string, keywords []KeywordInput) (int, error) {
	created := 0

	for i, kw := range keywords {
		if i > 0 && s.KeywordSearchDelay > 0 {
			select {
			case <-ctx.Done():
				return created, ctx.Err()
			case <-time.After(s.KeywordSearchDelay):
			}
		}

		posts, err := s.provider.Search(ctx, kw.Text)
		if err != nil {
			log.Printf("discovery search failed for keyword %q: %v", kw.Text, err)
			continue
		}

		if len(posts) > s.MaxPostsPerKeyword {
			posts = posts[:s.MaxPostsPerKeyword]
		}

		for _, post := range posts {
			if !scoring.PassesPrefilter(post) {
				continue
			}

			postID, err := s.repo.UpsertRedditPost(ctx, projectID, kw.ID, post)
			if err != nil {
				log.Printf("save reddit post %s failed: %v", post.ID, err)
				continue
			}

			exists, err := s.repo.OpportunityExists(ctx, projectID, postID)
			if err != nil {
				log.Printf("check existing opportunity failed: %v", err)
				continue
			}
			if exists {
				continue
			}

			llmResult, err := scoring.ClassifyPost(ctx, s.classifier, productDescription, post)
			if err != nil {
				log.Printf("classify post %s failed: %v", post.ID, err)
				continue
			}

			combined := scoring.Combine(scoring.RuleScore(post), llmResult)

			_, err = s.repo.Create(ctx, &Opportunity{
				ProjectID:      projectID,
				RedditPostID:   postID,
				RelevanceScore: combined.RelevanceScore,
				IntentScore:    combined.IntentScore,
				RuleScore:      combined.RuleScore,
				TotalScore:     combined.TotalScore,
				Reasoning:      combined.Reasoning,
				SuggestedReply: combined.SuggestedReply,
			})
			if err != nil {
				log.Printf("save opportunity for post %s failed: %v", post.ID, err)
				continue
			}
			created++
		}
	}

	return created, nil
}

// TopForProject returns the top-ranked opportunities for a project.
func (s *Service) TopForProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*Opportunity, error) {
	return s.repo.TopByProject(ctx, projectID, limit)
}
