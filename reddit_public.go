package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PublicProvider implements Provider by querying Reddit's public search JSON
// endpoint (https://www.reddit.com/search.json). This requires no OAuth
// app, no client_id/secret, and no login.
//
// Reddit blocks unauthenticated requests from many cloud/hosting IP ranges
// (Render, Fly, Railway, etc.) with a 403, regardless of headers. To work
// around that without requiring a Reddit developer app, this provider falls
// back to a small list of free public fetch proxies — each just forwards
// the request from its own IP, which usually isn't on Reddit's blocklist.
// This is best-effort: free proxies can be flaky or rate-limited, so
// several are tried in order before giving up. If a Reddit OAuth app is
// available, prefer OAuthProvider instead (see factory.go) — it's the
// sanctioned path and doesn't need any of this.
type PublicProvider struct {
	userAgent  string
	httpClient *http.Client

	// Limit caps how many posts are requested per search call.
	Limit int
}

func NewPublicProvider(userAgent string) *PublicProvider {
	if userAgent == "" {
		userAgent = "reddit-leads/0.1 (public search client)"
	}
	return &PublicProvider{
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		Limit:      25,
	}
}

// fetchAttempt is one way to reach the target URL: either direct, or via a
// proxy that wraps the target URL.
type fetchAttempt struct {
	name     string
	buildURL func(target string) string
}

var fetchAttempts = []fetchAttempt{
	{
		name:     "direct",
		buildURL: func(target string) string { return target },
	},
	{
		// https://api.allorigins.win — free, no signup, fetches server-side
		// and returns raw bytes.
		name: "allorigins",
		buildURL: func(target string) string {
			return "https://api.allorigins.win/raw?url=" + url.QueryEscape(target)
		},
	},
	{
		// https://corsproxy.io — free, no signup.
		name: "corsproxy.io",
		buildURL: func(target string) string {
			return "https://corsproxy.io/?url=" + url.QueryEscape(target)
		},
	},
	{
		// https://r.jina.ai — free reader proxy; passes raw JSON through
		// for API endpoints.
		name: "r.jina.ai",
		buildURL: func(target string) string {
			return "https://r.jina.ai/" + target
		},
	},
}

func (p *PublicProvider) Search(ctx context.Context, query string) ([]RedditPost, error) {
	target := fmt.Sprintf(
		"https://www.reddit.com/search.json?q=%s&sort=new&limit=%d&type=link",
		url.QueryEscape(query), p.Limit,
	)

	var lastErr error
	for _, attempt := range fetchAttempts {
		posts, err := p.tryFetch(ctx, attempt.buildURL(target))
		if err == nil {
			return posts, nil
		}
		lastErr = fmt.Errorf("%s: %w", attempt.name, err)
	}

	return nil, fmt.Errorf("all reddit fetch attempts failed, last error: %w", lastErr)
}

func (p *PublicProvider) tryFetch(ctx context.Context, endpoint string) ([]RedditPost, error) {
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
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate-limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB cap
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var payload redditListing
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
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
