package scoring

import (
	"strings"
	"time"

	"github.com/codersirojiddin/reddit-leads/internal/discovery"
)

// Weights control how the final total_score is composed from the rule-based
// score and the two LLM-derived scores (relevance, intent).
const (
	WeightRule      = 0.15
	WeightRelevance = 0.35
	WeightIntent    = 0.50

	// MaxPostAgeDays filters out posts older than this many days: old threads
	// are unlikely to yield a warm lead.
	MaxPostAgeDays = 30
	// MinBodyLength filters out posts that are too short to judge intent from.
	MinBodyLength = 20
)

// PassesPrefilter applies cheap, deterministic checks to skip posts before
// spending an LLM call on them.
func PassesPrefilter(p discovery.RedditPost) bool {
	if p.Title == "" {
		return false
	}

	combined := strings.ToLower(p.Title + " " + p.Body)
	if strings.Contains(combined, "[removed]") || strings.Contains(combined, "[deleted]") {
		return false
	}

	if len(p.Title)+len(p.Body) < MinBodyLength {
		return false
	}

	age := time.Since(time.Unix(p.CreatedAt, 0))
	if age > MaxPostAgeDays*24*time.Hour {
		return false
	}

	return true
}

// RuleScore computes a small deterministic score (0-100) from post
// engagement signals, used as a lightweight prior alongside the LLM scores.
func RuleScore(p discovery.RedditPost) float64 {
	score := 0.0

	// Engagement: more upvotes/comments suggest a real, visible conversation.
	switch {
	case p.Score >= 50:
		score += 40
	case p.Score >= 10:
		score += 25
	case p.Score >= 1:
		score += 10
	}

	switch {
	case p.NumComments >= 20:
		score += 30
	case p.NumComments >= 5:
		score += 20
	case p.NumComments >= 1:
		score += 10
	}

	// Recency: newer posts are more actionable.
	ageHours := time.Since(time.Unix(p.CreatedAt, 0)).Hours()
	switch {
	case ageHours <= 24:
		score += 30
	case ageHours <= 24*7:
		score += 20
	case ageHours <= 24*30:
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score
}
