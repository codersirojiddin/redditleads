# reddit-leads

Finds Reddit posts that look like real leads for your product: give it a
product description and up to 5 keywords, and it returns a ranked list of
posts worth replying to, with a suggested reply for each.

## Stack

- **Backend**: Go 1.24, `apps/api` (HTTP) + `apps/worker` (background jobs) sharing `internal/`
- **Frontend**: React + TypeScript + Vite + Tailwind (`web/`)
- **Database**: Postgres (via `pgx`)
- **Discovery**: Reddit API (OAuth2 client-credentials / "script" app)
- **Scoring**: OpenAI (or any compatible Chat Completions API) for relevance/intent scoring
- **Billing**: Paddle (planned, not yet wired — see `docs/architecture.md`)

## Free hosting deploy (no server, no domain — GitHub push → live)

This is the fastest path to a live URL with zero infrastructure cost or
setup. Three free pieces: **Supabase** (Postgres), **Render** (API + worker,
combined into one free web service), and **Render Static Site** or
**Netlify** (frontend).

### 1. Push the code to GitHub

```bash
cd reddit-leads
git init
git add .
git commit -m "Initial commit"
```
Create an empty repo on https://github.com/new, then:
```bash
git remote add origin https://github.com/YOUR_USERNAME/reddit-leads.git
git push -u origin main
```

### 2. Database — Supabase (free, permanent)

1. https://supabase.com → new project (pick any region, set a DB password)
2. Project Settings → Database → **Connection string** → copy the "URI"
   (use the **Session pooler** connection string, port 6543 — it handles
   many short-lived connections better than the direct one, which matters
   for a free-tier app). It looks like:
   ```
   postgresql://postgres.xxxxxxxx:[PASSWORD]@aws-0-xxxxx.pooler.supabase.com:6543/postgres
   ```
3. Keep this — it's your `DATABASE_URL`.

### 3. API + worker — Render Web Service (free)

1. https://render.com → New → **Web Service** → connect your GitHub repo
2. Settings:
   - **Runtime**: Docker
   - **Dockerfile Path**: `Dockerfile.render` (a dedicated single-service
     image — see the note below on why not the root `Dockerfile`)
   - **Docker Build Context Directory**: `.`
   - **Instance Type**: Free
3. Environment variables (Render dashboard → Environment):
   ```
   DATABASE_URL=<your Supabase connection string>
   JWT_SECRET=<openssl rand -hex 32>
   OPENAI_API_KEY=<your key>
   OPENAI_MODEL=gpt-4o-mini
   REDDIT_USER_AGENT=reddit-leads/0.1 by yourusername
   FRONTEND_URL=<filled in after step 4, e.g. https://reddit-leads.onrender.com>
   PORT=8080
   ```
   (`RUN_WORKER_INMEMORY` defaults to `true` — leave it unset. This makes the
   one free service run both the HTTP API and the background report worker.)
4. Deploy. Once live, copy its public URL (e.g.
   `https://reddit-leads-api.onrender.com`) — you'll need it for the frontend.

**Why `Dockerfile.render` and not the root `Dockerfile`?** The root
Dockerfile has two final stages (`api` and `worker`) for docker-compose,
which needs a specific target — but Render's dashboard doesn't offer a
"build target" field. Building it untargeted would produce the `worker`
image (no HTTP server) since it's defined last. `Dockerfile.render` is a
dedicated single-stage build that always produces the API-plus-worker image.

**Free-tier caveat**: Render's free web services spin down after ~15
minutes of no traffic and take 30-60s to wake up on the next request —
including the background worker, since it's the same process. Fine for
testing; upgrade to a paid instance ($7/mo) once you have real usage to
keep it always-on.

### 4. Frontend — Render Static Site (free)

1. Render dashboard → New → **Static Site** → same GitHub repo
2. Settings:
   - **Root Directory**: `web`
   - **Build Command**: `npm install && npm run build`
   - **Publish Directory**: `dist`
3. Environment variable:
   ```
   VITE_API_BASE_URL=https://reddit-leads-api.onrender.com/api/v1
   ```
   (use the actual API URL from step 3)
4. Deploy. Copy this site's URL (e.g. `https://reddit-leads.onrender.com`).

### 5. Close the loop

Go back to the API service's environment variables (step 3) and set
`FRONTEND_URL` to the static site's URL from step 4, then redeploy the API
service — this is what allows the browser to call the API across the two
different `onrender.com` domains (CORS).

You're live. Register an account at the static site URL and go.

## Production deploy (VPS + domain, free HTTPS)

This repo includes a `caddy` service in `docker-compose.yml` that gets and
renews a free Let's Encrypt certificate automatically — no manual
nginx/certbot setup.

1. Point your domain's DNS **A record** at your server's IP (e.g. in your
   registrar's dashboard: `dunnet.online → A → <server IP>`). Wait for it to
   propagate (`dig +short yourdomain.com` should return the IP).
2. On the server: `cp .env.example .env`, fill in `JWT_SECRET`,
   `OPENAI_API_KEY`, and set:
   ```
   DOMAIN=yourdomain.com
   FRONTEND_URL=https://yourdomain.com
   ```
3. Open firewall ports 80 and 443 (443 in addition to 80 — Caddy needs both
   for the certificate challenge and HTTPS itself):
   ```bash
   ufw allow 80
   ufw allow 443
   ufw allow 22
   ufw enable
   ```
4. `docker compose up --build -d`

Caddy issues the certificate on first request — visiting `https://yourdomain.com`
right after startup may take a few seconds the very first time. Check
`docker compose logs -f caddy` if it doesn't come up.

## Quickstart (Docker, easiest)

```bash
cp .env.example .env
# fill in JWT_SECRET, OPENAI_API_KEY (REDDIT_USER_AGENT already has a default)

docker compose up --build
```

- Frontend: http://localhost:3000
- API: http://localhost:8080
- Optionally run `./scripts/seed.sh` afterwards to create a demo account
  with a project and keywords already set up.

## Quickstart (local dev, hot reload)

```bash
cp .env.example .env
# fill in secrets

./scripts/dev.sh
```

This starts Postgres in Docker, the Go API + worker with `go run`, and the
Vite dev server (with hot reload) — all in one command. Frontend at
http://localhost:3000, proxying `/api` to the Go API on :8080.

Or run each piece manually:
```bash
docker compose up -d db
go run ./apps/api      # auto-runs migrations, serves :8080
go run ./apps/worker    # separate process, polls for pending reports
cd web && npm install && npm run dev   # serves :3000
```

## Docs

- [`docs/architecture.md`](docs/architecture.md) — how the pieces fit together
- [`docs/scoring.md`](docs/scoring.md) — how `total_score` is computed
- [`docs/api.md`](docs/api.md) — full HTTP API reference

## Getting an OpenAI key

- OpenAI: https://platform.openai.com/api-keys → `OPENAI_API_KEY`.
  `OPENAI_MODEL` defaults to `gpt-4o-mini`.
- Reddit needs no credentials — discovery uses the public search endpoint
  (see `docs/architecture.md`). Just set `REDDIT_USER_AGENT` to something
  identifying yourself in `.env`.

## Project layout

```
apps/api/        Go HTTP server entrypoint
apps/worker/      Go background worker entrypoint
internal/         Go domain packages (auth, projects, keywords, discovery,
                   ai, scoring, opportunities, reports, jobs, http, config,
                   database, users)
migrations/        SQL schema, applied automatically on boot
web/               React + TypeScript + Tailwind frontend (Vite)
docs/              Architecture, scoring, and API reference docs
scripts/           dev.sh (local dev) and seed.sh (demo data)
Caddyfile          Reverse proxy + automatic HTTPS config for production (VPS deploy)
Dockerfile.render  Single-service image for free-tier hosts (Render, etc.)
```
