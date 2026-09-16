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

// redditListing and redditPostData mirror the shape of Reddit's JSON listing
// responses, shared by both PublicProvider (reddit.com/search.json, direct
// or via a fetch proxy) and OAuthProvider (oauth.reddit.com/search) since
// they return identical data.
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
