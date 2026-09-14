#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Running API smoke test against ${BASE_URL}"

echo "1. Checking health..."
curl --fail --silent --show-error \
    "${BASE_URL}/health" \
    > /dev/null

echo "2. Resolving seeded author..."

AUTHOR_ID="$(
    docker compose exec -T postgres \
        psql \
        -At \
        -U "${POSTGRES_USER}" \
        -d "${POSTGRES_DB}" \
        -c "SELECT id FROM authors WHERE name = 'J.R.R. Tolkien' LIMIT 1;"
)"

if [ -z "${AUTHOR_ID}" ]; then
    echo "Seeded author not found. Run 'make seed' first."
    exit 1
fi

TITLE="Smoke Test Book $(date +%s)"
UPDATED_TITLE="${TITLE} Updated"

HEADERS="$(mktemp)"
BODY="$(mktemp)"

cleanup() {
    rm -f "${HEADERS}" "${BODY}"
}

trap cleanup EXIT

echo "3. Creating book..."

curl --fail --silent --show-error \
    -D "${HEADERS}" \
    -o "${BODY}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "{
        \"title\": \"${TITLE}\",
        \"description\": \"Created by the API smoke test.\",
        \"publication_date\": \"2026-09-14\",
        \"author_ids\": [${AUTHOR_ID}]
    }" \
    "${BASE_URL}/v1/books"

LOCATION="$(
    awk '
        BEGIN { IGNORECASE=1 }
        /^Location:/ {
            gsub("\r", "", $2)
            print $2
        }
    ' "${HEADERS}"
)"

if [ -z "${LOCATION}" ]; then
    echo "Create response did not contain a Location header."
    exit 1
fi

BOOK_ID="${LOCATION##*/}"

echo "Created book ${BOOK_ID}"

echo "4. Retrieving book..."

curl --fail --silent --show-error \
    "${BASE_URL}/v1/books/${BOOK_ID}/" \
    > /dev/null

echo "5. Updating book..."

curl --fail --silent --show-error \
    -X PATCH \
    -H "Content-Type: application/json" \
    -d "{
        \"title\": \"${UPDATED_TITLE}\"
    }" \
    "${BASE_URL}/v1/books/${BOOK_ID}/" \
    > /dev/null

echo "6. Retrieving history..."

curl --fail --silent --show-error \
    "${BASE_URL}/v1/books/${BOOK_ID}/history" \
    > /dev/null

echo "7. Listing books..."

curl --fail --silent --show-error \
    "${BASE_URL}/v1/books?title=Smoke%20Test" \
    > /dev/null

echo "8. Deleting book..."

curl --fail --silent --show-error \
    -X DELETE \
    "${BASE_URL}/v1/books/${BOOK_ID}/" \
    > /dev/null

echo "9. Verifying deleted book is unavailable..."

STATUS="$(
    curl \
        --silent \
        --output /dev/null \
        --write-out "%{http_code}" \
        "${BASE_URL}/v1/books/${BOOK_ID}/"
)"

if [ "${STATUS}" != "404" ]; then
    echo "Expected deleted book to return 404, got ${STATUS}"
    exit 1
fi

echo "10. Verifying history survives deletion..."

curl --fail --silent --show-error \
    "${BASE_URL}/v1/books/${BOOK_ID}/history" \
    > /dev/null

echo "Smoke test passed."