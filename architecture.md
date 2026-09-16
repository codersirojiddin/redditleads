# Architecture

Two binaries share one Go module and one Postgres database:

- `apps/api` — HTTP API (chi router). Auth, projects, keywords, reports (request + read).
- `apps/worker` — polls the `reports` table for `status = pending` rows and runs
  the discovery → scoring → opportunities pipeline for each one.

There's no separate broker (Redis/SQS): a "job" is simply a `reports` row.
`ClaimNextPending` uses `SELECT ... FOR UPDATE SKIP LOCKED` so you can safely
run more than one worker instance without double-processing a report.

## Request flow

1. User registers/logs in → JWT (`internal/auth`).
2. User creates a project with a product description (`internal/projects`).
3. User adds up to 5 keywords (`internal/keywords`).
4. User requests a report: `POST /projects/{id}/reports` inserts a
   `reports` row with `status=pending` and returns immediately (202).
5. The worker picks it up:
   - `internal/discovery` searches Reddit for each keyword.
   - `internal/scoring` cheaply prefilters junk posts (removed/deleted, too old, too short).
   - `internal/ai` sends each surviving post + the product description to an
     LLM, asking for a relevance score, an intent score, reasoning, and a
     suggested reply — all as structured JSON.
   - `internal/scoring` combines the rule score with the LLM scores into one
     `total_score` (see `docs/scoring.md`).
   - `internal/opportunities` persists each scored post as an `opportunities` row.
   - The top 100 opportunities for the project are linked into `report_items`,
     and the report is marked `completed`.
6. User polls or is notified, then reads `GET /reports/{id}` for the ranked list.

## Why an AI/LLM API is needed

Reddit's search API only does keyword matching — it can't tell a genuine
buying-intent post ("is there a lighter alternative to X?") from a post that
merely contains the keyword by coincidence. The LLM call is what turns raw
keyword matches into qualified leads (relevance + intent scoring, plus a
ready-to-post reply suggestion). See `internal/ai/prompts.go`.

## Reddit discovery: two providers, picked automatically

`internal/discovery` has two implementations of the `Provider` interface,
chosen by `NewProvider` (see `factory.go`) based on whether Reddit app
credentials are configured:

- **`PublicProvider`** (default, no setup): queries Reddit's public search
  endpoint (`reddit.com/search.json`) — the same data a logged-out browser
  sees. No Reddit app, client ID/secret, or login needed. Reddit blocks
  unauthenticated requests from many cloud/hosting IP ranges (Render, Fly,
  etc.) with a 403 — to work around that without a developer app, this
  provider falls back through a short list of free public fetch proxies
  (allorigins, corsproxy.io, r.jina.ai) that make the request from their own
  IP. Best-effort: free proxies can be flaky, so this isn't as reliable as
  the OAuth path, but needs zero setup.
- **`OAuthProvider`** (used automatically once configured): queries the
  authenticated API (`oauth.reddit.com`) using a Reddit "script" app's
  client-credentials token. This is Reddit's sanctioned API path and isn't
  subject to the same IP blocking. Set `REDDIT_CLIENT_ID` and
  `REDDIT_CLIENT_SECRET` (from https://www.reddit.com/prefs/apps) to switch
  to it — no code changes needed.

## Data model

```
users 1---* projects 1---* keywords
                    1---* reddit_posts
                    1---* opportunities  (one per reddit_post, scored)
                    1---* reports 1---* report_items ---> opportunities
```

## Not yet wired (intentionally out of scope for today)

- Paddle billing (subscriptions/usage limits) — env vars are scaffolded in
  `.env.example` but there's no `internal/billing` package yet. Wire it in
  once the core lead-finding loop is validated.
- Email delivery of reports (Resend) — `EMAIL_FROM` / `RESEND_API_KEY` are
  scaffolded; add an `internal/notify` package that reads a completed report
  and sends it.
