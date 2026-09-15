package discovery

import (
	"context"
	"time"
)

// KeywordResult pairs a keyword with the posts found for it.
type KeywordResult struct {
	Keyword string
	Posts   []RedditPost
	Err     error
}

// SearchKeywords runs provider.Search for each keyword sequentially, with a
// small delay between calls to stay well under Reddit's rate limits. A
// failure on one keyword does not stop the others; the error is attached to
// that keyword's result instead.
func SearchKeywords(ctx context.Context, provider Provider, keywords []string, delayBetweenCalls time.Duration) []KeywordResult {
	results := make([]KeywordResult, 0, len(keywords))

	for i, kw := range keywords {
		posts, err := provider.Search(ctx, kw)
		results = append(results, KeywordResult{Keyword: kw, Posts: posts, Err: err})

		if i < len(keywords)-1 && delayBetweenCalls > 0 {
			select {
			case <-ctx.Done():
				return results
			case <-time.After(delayBetweenCalls):
			}
		}
	}

	return results
}
