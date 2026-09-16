package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthProvider implements Provider by querying Reddit's authenticated
// search API (oauth.reddit.com) using an app-only OAuth2 "client
// credentials" token from a Reddit "script" app
// (https://www.reddit.com/prefs/apps). Unlike PublicProvider, requests go
// through Reddit's sanctioned API path, which is not subject to the same
// anti-scraping IP blocking that affects unauthenticated requests to
// reddit.com from cloud/hosting IP ranges — the preferred provider once you
// have app credentials (see factory.go).
type OAuthProvider struct {
	clientID     string
	clientSecret string
	userAgent    string
	httpClient   *http.Client

	// Limit caps how many posts are requested per search call.
	Limit int

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewOAuthProvider(clientID, clientSecret, userAgent string) *OAuthProvider {
	if userAgent == "" {
		userAgent = "reddit-leads/0.1 (oauth client)"
	}
	return &OAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		userAgent:    userAgent,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		Limit:        25,
	}
}

func (p *OAuthProvider) Search(ctx context.Context, query string) ([]RedditPost, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("reddit auth: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://oauth.reddit.com/search?q=%s&sort=new&limit=%d&type=link",
		url.QueryEscape(query), p.Limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit search request: %w", err)
	}
	defer resp.Body.Close()

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

// getToken returns a cached app-only access token, refreshing it if expired.
func (p *OAuthProvider) getToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://www.reddit.com/api/v1/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(p.clientID, p.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reddit token endpoint returned status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	p.accessToken = tokenResp.AccessToken
	// Refresh a little early to avoid edge-of-expiry failures.
	p.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return p.accessToken, nil
}
