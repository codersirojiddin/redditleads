package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ClassificationResult is the structured judgement returned for a single post.
type ClassificationResult struct {
	RelevanceScore int    `json:"relevance_score"`
	IntentScore    int    `json:"intent_score"`
	Reasoning      string `json:"reasoning"`
	SuggestedReply string `json:"suggested_reply"`
}

type Classifier struct {
	client *Client
}

func NewClassifier(client *Client) *Classifier {
	return &Classifier{client: client}
}

// Classify judges a single Reddit post against a product description.
func (c *Classifier) Classify(ctx context.Context, productDescription, postTitle, postBody, subreddit string) (*ClassificationResult, error) {
	userPrompt := BuildClassifierUserPrompt(productDescription, postTitle, postBody, subreddit)

	raw, err := c.client.CompleteJSON(ctx, ClassifierPromptV1, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("classify post: %w", err)
	}

	raw = strings.TrimSpace(raw)
	var result ClassificationResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse classifier response: %w (raw: %s)", err, truncate(raw, 300))
	}

	result.RelevanceScore = clamp(result.RelevanceScore, 0, 100)
	result.IntentScore = clamp(result.IntentScore, 0, 100)

	return &result, nil
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
