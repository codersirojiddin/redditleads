#!/usr/bin/env bash
# Creates a demo user, project, and keywords via the running API so you have
# something to look at immediately. Requires the API to be running (see
# scripts/dev.sh) and `jq` installed.
set -euo pipefail

API="${API_URL:-http://localhost:8080/api/v1}"
EMAIL="${SEED_EMAIL:-demo@example.com}"
PASSWORD="${SEED_PASSWORD:-password123}"

echo "Registering demo user ($EMAIL)..."
TOKEN=$(curl -s -X POST "$API/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.token // empty')

if [ -z "$TOKEN" ]; then
  echo "Register failed (maybe already exists) — trying login..."
  TOKEN=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.token')
fi

echo "Creating demo project..."
PROJECT_ID=$(curl -s -X POST "$API/projects/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Notely","product_description":"A lightweight note-taking app for people who find Notion too heavy and just want fast, distraction-free notes.","target_url":"https://example.com"}' \
  | jq -r '.id')

echo "Adding keywords..."
for kw in "notion alternative" "lightweight note taking app" "distraction free notes"; do
  curl -s -X POST "$API/projects/$PROJECT_ID/keywords/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"keyword\":\"$kw\"}" > /dev/null
done

echo "Done. Log in at http://localhost:3000 with:"
echo "  email:    $EMAIL"
echo "  password: $PASSWORD"
echo "Then click 'Find new leads' on the Notely project to run a report."
