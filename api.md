# API

Base path: `/api/v1`. All authenticated routes require `Authorization: Bearer <jwt>`.

## Auth

### POST /auth/register
```json
{ "email": "you@example.com", "password": "at-least-8-chars" }
```
→ `201` `{ "token": "...", "user": { "id": "...", "email": "..." } }`

### POST /auth/login
Same body shape as register → `200` with the same response shape.

## Me

### GET /me *(auth required)*
→ `200` `{ "id": "...", "email": "..." }`

## Projects

### POST /projects *(auth required)*
```json
{ "name": "My SaaS", "product_description": "A tool that...", "target_url": "https://..." }
```
→ `201` project object

### GET /projects *(auth required)*
→ `200` array of project objects, newest first

### GET /projects/{projectID} *(auth required)*
### DELETE /projects/{projectID} *(auth required)* → `204`

## Keywords

Max 5 keywords per project (`400` once the limit is reached).

### POST /projects/{projectID}/keywords *(auth required)*
```json
{ "keyword": "note taking app alternative" }
```
→ `201` `{ "id": "...", "keyword": "..." }`

### GET /projects/{projectID}/keywords *(auth required)*
### DELETE /projects/{projectID}/keywords/{keywordID} *(auth required)* → `204`

## Reports

### POST /projects/{projectID}/reports *(auth required)*
Enqueues a new discovery+scoring run (picked up by the worker).
→ `202` `{ "id": "...", "status": "pending", ... }`

### GET /projects/{projectID}/reports *(auth required)*
→ `200` array of report objects (without opportunities — use GET by id for that)

### GET /reports/{reportID} *(auth required)*
→ `200`:
```json
{
  "id": "...",
  "project_id": "...",
  "status": "completed",
  "created_at": "...",
  "completed_at": "...",
  "opportunities": [
    {
      "id": "...",
      "post_title": "...",
      "post_url": "https://www.reddit.com/...",
      "subreddit": "...",
      "author": "...",
      "relevance_score": 82,
      "intent_score": 91,
      "total_score": 86.7,
      "reasoning": "...",
      "suggested_reply": "..."
    }
  ]
}
```

Report status values: `pending → running → completed | failed`. Poll this
endpoint (or add a webhook/email later) until status leaves `pending`/`running`.
