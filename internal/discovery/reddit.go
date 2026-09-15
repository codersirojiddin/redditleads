package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RedditProvider implements Provider by querying Reddit's public search JSON
// endpoint (https://www.reddit.com/search.json). This requires no OAuth app,
// no client_id/secret, and no login — it's the same data a logged-out
// browser sees. Trade-off: Reddit rate-limits this endpoint more
// aggressively than the authenticated API (roughly ~10 requests/minute per
// IP), and may occasionally block a server IP outright. That's an
// acceptable trade for shipping today; if it becomes a bottleneck later,
// swap this for an OAuth-based provider — both satisfy the same Provider
// interface, so nothing else in the codebase needs to change.
type RedditProvider struct {
	userAgent  string
	httpClient *http.Client

	// Limit caps how many posts are requested per search call.
	Limit int
}

func NewRedditProvider(userAgent string) *RedditProvider {
	if userAgent == "" {
		userAgent = "reddit-leads/0.1 (public search client)"
	}
	return &RedditProvider{
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		Limit:      25,
	}
}

func (p *RedditProvider) Search(ctx context.Context, query string) ([]RedditPost, error) {
	endpoint := fmt.Sprintf(
		"https://www.reddit.com/search.json?q=%s&sort=new&limit=%d&type=link",
		url.QueryEscape(query), p.Limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	// A descriptive, non-generic User-Agent matters a lot here: Reddit
	// blocks generic/default user agents (e.g. Go's default) much faster
	// than identifiable ones.
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("reddit rate-limited this request (429) — back off and retry later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit search returned status %d", resp.StatusCode)
	}

	var payload redditListing
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode reddit response: %w", err)
	}

	posts := make([]RedditPost, 0, len(payload.Data.Children))
	for _, child := range payload.Data.Children {
		d := child.Data
		if d.RemovedByCategory != "" || d.Selftext == "[removed]" || d.Selftext == "[deleted]" {
			continue
		}
		posts = append(posts, RedditPost{
			ID:          d.ID,
			Title:       d.Title,
			Body:        d.Selftext,
			Subreddit:   d.Subreddit,
			Author:      d.Author,
			URL:         "https://www.reddit.com" + d.Permalink,
			CreatedAt:   int64(d.CreatedUTC),
			Score:       d.Score,
			NumComments: d.NumComments,
		})
	}

	return posts, nil
}

type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPostData `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPostData struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Selftext          string  `json:"selftext"`
	Subreddit         string  `json:"subreddit"`
	Author            string  `json:"author"`
	Permalink         string  `json:"permalink"`
	CreatedUTC        float64 `json:"created_utc"`
	Score             int     `json:"score"`
	NumComments       int     `json:"num_comments"`
	RemovedByCategory string  `json:"removed_by_category"`
}
