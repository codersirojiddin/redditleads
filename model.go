package opportunities

import (
	"time"

	"github.com/google/uuid"
)

type Opportunity struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	RedditPostID   uuid.UUID
	RelevanceScore float64
	IntentScore    float64
	RuleScore      float64
	TotalScore     float64
	Reasoning      string
	SuggestedReply string
	Status         string
	CreatedAt      time.Time

	// Populated by joined queries (not stored on this table).
	PostTitle     string
	PostURL       string
	PostSubreddit string
	PostAuthor    string
}
