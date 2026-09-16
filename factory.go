package discovery

// NewProvider picks the best available Reddit discovery provider:
//   - If Reddit "script" app credentials are configured, uses OAuthProvider
//     (oauth.reddit.com) — the sanctioned path, not subject to the IP
//     blocking that affects unauthenticated requests.
//   - Otherwise falls back to PublicProvider (reddit.com/search.json, with
//     free fetch-proxy fallbacks for blocked hosting IPs) — no credentials
//     needed.
func NewProvider(clientID, clientSecret, userAgent string) Provider {
	if clientID != "" && clientSecret != "" {
		return NewOAuthProvider(clientID, clientSecret, userAgent)
	}
	return NewPublicProvider(userAgent)
}
