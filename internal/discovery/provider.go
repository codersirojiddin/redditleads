package discovery

import "context"

// RedditPost is a normalized post returned by any discovery Provider.
type RedditPost struct {
	ID          string
	Title       string
	Body        string
	Subreddit   string
	Author      string
	URL         string
	CreatedAt   int64
	Score       int
	NumComments int
}

// Provider abstracts a source of Reddit posts for a search query, so the
// concrete implementation (real Reddit API, a mock, etc.) can be swapped.
type Provider interface {
	Search(ctx context.Context, query string) ([]RedditPost, error)
}
