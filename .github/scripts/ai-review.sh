#!/bin/bash
set -euo pipefail

# AI code review against the ThreatWinds /chat/completions endpoint.
#
# Reads a prompt file (with optional YAML frontmatter that can override the
# model), appends the PR diff, calls the model, and writes the parsed JSON
# verdict to OUTPUT_FILE. Does NOT post PR comments — the approver job
# consolidates all prompt results and decides what to comment.
#
# Always exits 0 (data-producer role). If parsing fails the output JSON has
# tier=2 with a generic "review manually" finding (fail-safe).
#
# Required env vars:
#   PROMPT_FILE              path to a .github/ai-prompts/*.md file
#   DIFF_FILE                path to a unified diff
#   OUTPUT_FILE              where to write the JSON result
#   THREATWINDS_API_KEY      auth
#   THREATWINDS_API_SECRET   auth
#
# Optional env vars:
#   AI_REVIEW_MODEL          default model when the prompt doesn't pin one
#   THREATWINDS_BASE_URL     defaults to https://apis.threatwinds.com/api/ai/v1
#   MAX_DIFF_BYTES           truncate the diff above this size (default 200000)

: "${PROMPT_FILE:?PROMPT_FILE is required}"
: "${DIFF_FILE:?DIFF_FILE is required}"
: "${OUTPUT_FILE:?OUTPUT_FILE is required}"
: "${THREATWINDS_API_KEY:?THREATWINDS_API_KEY is required}"
: "${THREATWINDS_API_SECRET:?THREATWINDS_API_SECRET is required}"

DEFAULT_MODEL="${AI_REVIEW_MODEL:-gemini-3-flash-lite}"
BASE_URL="${THREATWINDS_BASE_URL:-https://apis.threatwinds.com/api/ai/v1}"
MAX_DIFF_BYTES="${MAX_DIFF_BYTES:-200000}"

# --- Helper: write a fallback result and exit 0 (always succeed) -------------

write_fallback() {
    local reason="$1"
    jq -n \
        --arg prompt "$prompt_name" \
        --arg model "$MODEL" \
        --arg reason "$reason" \
        '{
            prompt: $prompt,
            model: $model,
            tier: 2,
            summary: "AI review could not parse model response — manual review recommended.",
            findings: [{
                severity: "high",
                file: "(n/a)",
                line: 0,
                message: ($reason + " (fail-safe: a review that cannot run is treated as blocking).")
            }]
        }' > "$OUTPUT_FILE"
    echo "::warning::Wrote fallback result: $reason"
    exit 0
}

# --- Parse frontmatter -------------------------------------------------------

prompt_name="$(basename "$PROMPT_FILE" .md)"
prompt_model=""
body_start=1

if head -n 1 "$PROMPT_FILE" | grep -qx -- '---'; then
    end_line=$(awk 'NR>1 && /^---$/ {print NR; exit}' "$PROMPT_FILE")
    if [[ -n "$end_line" ]]; then
        while IFS= read -r line; do
            key="${line%%:*}"
            value="${line#*:}"
            value="$(echo "$value" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
            case "$key" in
                name)  prompt_name="$value" ;;
                model) prompt_model="$value" ;;
            esac
        done < <(sed -n "2,$((end_line - 1))p" "$PROMPT_FILE")
        body_start=$((end_line + 1))
    fi
fi

MODEL="${prompt_model:-$DEFAULT_MODEL}"

echo "::group::AI review — prompt: $prompt_name (model: $MODEL)"

# --- Nothing to review -------------------------------------------------------
# The diff can be empty after upstream filtering (e.g. a PR that only touches
# excluded rules/filters/definitions paths). Pass as Tier 1 instead of calling
# the model with an empty diff.
if [[ ! -s "$DIFF_FILE" ]] || ! grep -q '[^[:space:]]' "$DIFF_FILE"; then
    jq -n --arg prompt "$prompt_name" --arg model "$MODEL" \
        '{prompt: $prompt, model: $model, tier: 1, summary: "No reviewable changes in this diff (excluded paths only).", findings: []}' \
        > "$OUTPUT_FILE"
    echo "Empty diff — wrote Tier 1 pass."
    echo "::endgroup::"
    exit 0
fi

# --- Build request body ------------------------------------------------------

prompt_body=$(tail -n "+${body_start}" "$PROMPT_FILE")

diff_bytes=$(wc -c < "$DIFF_FILE" | tr -d ' ')
if (( diff_bytes > MAX_DIFF_BYTES )); then
    diff_content=$(head -c "$MAX_DIFF_BYTES" "$DIFF_FILE")
    diff_content+=$'\n\n[diff truncated: original '"$diff_bytes"' bytes, kept first '"$MAX_DIFF_BYTES"']'
else
    diff_content=$(cat "$DIFF_FILE")
fi

# Write the user message to a temp file. Passing it through --arg would hit
# the system ARG_MAX limit on PRs with large diffs ("Argument list too long").
user_message_file=$(mktemp)
printf '%s\n\n---\n\nPR diff to review:\n\n```diff\n%s\n```\n' \
    "$prompt_body" "$diff_content" > "$user_message_file"

request_body_file=$(mktemp)
jq -n \
    --arg model "$MODEL" \
    --rawfile content "$user_message_file" \
    '{
        model: $model,
        messages: [{role: "user", content: $content}],
        temperature: 0.2
    }' > "$request_body_file"

# --- Call the API ------------------------------------------------------------

response_file=$(mktemp)
http_status=$(curl -sS -o "$response_file" -w '%{http_code}' \
    -X POST "${BASE_URL%/}/chat/completions" \
    -H "Content-Type: application/json" \
    -H "api-key: ${THREATWINDS_API_KEY}" \
    -H "api-secret: ${THREATWINDS_API_SECRET}" \
    --data-binary "@${request_body_file}" || echo "000")

if [[ "$http_status" != "200" ]]; then
    echo "ThreatWinds API HTTP $http_status"
    cat "$response_file"
    echo "::endgroup::"
    write_fallback "ThreatWinds API returned HTTP $http_status"
fi

content=$(jq -r '.choices[0].message.content // empty' "$response_file")
if [[ -z "$content" ]]; then
    echo "Empty content from model"
    cat "$response_file"
    echo "::endgroup::"
    write_fallback "Model returned empty content"
fi

# Strip optional ```json fences.
content_clean=$(echo "$content" | sed -E 's/^```(json)?//; s/```$//' | sed '/^[[:space:]]*$/d')

if ! verdict_json=$(echo "$content_clean" | jq -c '.' 2>/dev/null); then
    echo "Model did not return valid JSON. Raw content:"
    echo "$content"
    echo "::endgroup::"
    write_fallback "Model response was not valid JSON"
fi

# Validate schema minimally: tier must be 1/2/3.
tier=$(echo "$verdict_json" | jq -r '.tier // 0')
if [[ "$tier" != "1" && "$tier" != "2" && "$tier" != "3" ]]; then
    echo "Invalid tier in response: $tier"
    echo "$verdict_json"
    echo "::endgroup::"
    write_fallback "Model returned invalid tier value (expected 1, 2, or 3)"
fi

# --- Persist enriched result -------------------------------------------------
# Inject prompt + model identifiers so the approver doesn't need to know them.
final=$(echo "$verdict_json" | jq \
    --arg prompt "$prompt_name" \
    --arg model "$MODEL" \
    '. + {prompt: $prompt, model: $model}')

echo "$final" > "$OUTPUT_FILE"

summary=$(echo "$final" | jq -r '.summary // "(no summary)"')
findings_count=$(echo "$final" | jq -r '.findings | length // 0')
echo "Tier:     $tier"
echo "Summary:  $summary"
echo "Findings: $findings_count"
echo "::endgroup::"

exit 0
