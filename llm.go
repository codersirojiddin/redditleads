package scoring

import (
	"context"

	"github.com/codersirojiddin/reddit-leads/internal/ai"
	"github.com/codersirojiddin/reddit-leads/internal/discovery"
)

// LLMResult mirrors ai.ClassificationResult to keep the scoring package
// decoupled from the ai package's exact type (small, but keeps layering clean).
type LLMResult struct {
	RelevanceScore int
	IntentScore    int
	Reasoning      string
	SuggestedReply string
}

// ClassifyPost calls the AI classifier for a single post and adapts the result.
func ClassifyPost(ctx context.Context, classifier *ai.Classifier, productDescription string, post discovery.RedditPost) (*LLMResult, error) {
	result, err := classifier.Classify(ctx, productDescription, post.Title, post.Body, post.Subreddit)
	if err != nil {
		return nil, err
	}

	return &LLMResult{
		RelevanceScore: result.RelevanceScore,
		IntentScore:    result.IntentScore,
		Reasoning:      result.Reasoning,
		SuggestedReply: result.SuggestedReply,
	}, nil
}
