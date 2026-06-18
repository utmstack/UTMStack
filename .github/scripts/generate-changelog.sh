#!/usr/bin/env bash
#
# Generates AI-powered release notes via ThreatWinds /chat/completions.
# Used by .github/workflows/generate-changelog.yml; also runnable locally
# (export THREATWINDS_API_KEY + THREATWINDS_API_SECRET and run the script).
#
# Usage:
#   bash .github/scripts/generate-changelog.sh <current_tag> [previous_tag]
#
# Examples:
#   bash .github/scripts/generate-changelog.sh v11.2.8 v11.2.7
#   bash .github/scripts/generate-changelog.sh v11.2.8           # auto-detect previous
#
# Required env vars:
#   THREATWINDS_API_KEY      auth
#   THREATWINDS_API_SECRET   auth
#
# Optional env vars:
#   PRODUCT_NAME             default: "UTMStack"
#   PRODUCT_DESCRIPTION      default: "Unified Threat Management and SIEM Platform"
#   MODEL                    default: "gemini-3-flash-lite"
#   TEMPERATURE              default: 0.3
#   MAX_TOKENS               default: 2000
#   OUTPUT_FILE              default: "/tmp/changelog.md"
#   THREATWINDS_BASE_URL     default: https://apis.threatwinds.com/api/ai/v1

set -euo pipefail

# ─── Config ───────────────────────────────────────────────────────────────────
PRODUCT_NAME="${PRODUCT_NAME:-UTMStack}"
PRODUCT_DESCRIPTION="${PRODUCT_DESCRIPTION:-Unified Threat Management and SIEM Platform}"
MODEL="${MODEL:-gemini-3-flash-lite}"
TEMPERATURE="${TEMPERATURE:-0.3}"
MAX_TOKENS="${MAX_TOKENS:-2000}"
OUTPUT_FILE="${OUTPUT_FILE:-/tmp/changelog.md}"
BASE_URL="${THREATWINDS_BASE_URL:-https://apis.threatwinds.com/api/ai/v1}"

# ─── Arg parsing ──────────────────────────────────────────────────────────────
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <current_tag> [previous_tag]" >&2
    exit 1
fi

CURRENT_TAG="$1"
PREVIOUS_TAG="${2:-}"

# ─── Dependencies ─────────────────────────────────────────────────────────────
command -v jq   >/dev/null || { echo "jq is required";   exit 1; }
command -v curl >/dev/null || { echo "curl is required"; exit 1; }
command -v git  >/dev/null || { echo "git is required";  exit 1; }

: "${THREATWINDS_API_KEY:?THREATWINDS_API_KEY is required}"
: "${THREATWINDS_API_SECRET:?THREATWINDS_API_SECRET is required}"

# ─── Resolve current ref (tag may not exist yet when pipeline runs) ────────────
if git rev-parse "$CURRENT_TAG" >/dev/null 2>&1; then
    CURRENT_REF="$CURRENT_TAG"
else
    echo "Tag $CURRENT_TAG not found in repo yet, using HEAD"
    CURRENT_REF="HEAD"
fi

# ─── Resolve previous tag if not provided ─────────────────────────────────────
if [ -z "$PREVIOUS_TAG" ]; then
    echo "Auto-detecting previous tag..."
    ALL_TAGS=$(git tag --sort=-v:refname)
    if [ "$CURRENT_REF" = "$CURRENT_TAG" ]; then
        # Tag exists: find the tag immediately before CURRENT_TAG
        FOUND_CURRENT=false
        for tag in $ALL_TAGS; do
            if [ "$FOUND_CURRENT" = true ]; then
                PREVIOUS_TAG="$tag"
                break
            fi
            if [ "$tag" = "$CURRENT_TAG" ]; then
                FOUND_CURRENT=true
            fi
        done
    else
        # Tag doesn't exist yet: use the most recent existing tag
        PREVIOUS_TAG=$(echo "$ALL_TAGS" | head -1)
        [ -n "$PREVIOUS_TAG" ] && echo "Tag not yet created; using most recent existing tag: $PREVIOUS_TAG"
    fi
    if [ -z "$PREVIOUS_TAG" ]; then
        PREVIOUS_TAG=$(git rev-list --max-parents=0 HEAD | head -1)
        echo "No previous tag found, using first commit: $PREVIOUS_TAG"
    fi
fi

echo "Current tag:  $CURRENT_TAG"
echo "Current ref:  $CURRENT_REF"
echo "Previous tag: $PREVIOUS_TAG"
echo "Model:        $MODEL"
echo

# ─── Collect commits ──────────────────────────────────────────────────────────
COMMITS=$(git log "${PREVIOUS_TAG}..${CURRENT_REF}" --pretty=format:"- %h %s (%an)" --no-merges)
COMMIT_COUNT=$(git rev-list --count "${PREVIOUS_TAG}..${CURRENT_REF}" --no-merges)

if [ -z "$COMMITS" ]; then
    echo "No commits found between $PREVIOUS_TAG and $CURRENT_TAG."
    exit 0
fi

echo "Found $COMMIT_COUNT commits."
echo

# ─── Build prompt ─────────────────────────────────────────────────────────────
PROMPT="You are a product marketing writer creating release notes for end users of a software product.

Product: $PRODUCT_NAME - $PRODUCT_DESCRIPTION
Release: $CURRENT_TAG

Here are the commit messages from this release:
$COMMITS

Create user-friendly release notes in markdown format. This is for NON-TECHNICAL end users who want to know what's new and improved in the product.

IMPORTANT RULES:
1. ONLY include changes that DIRECTLY AFFECT END USERS - things they can see, use, or benefit from
2. COMPLETELY IGNORE internal/technical changes like:
   - CI/CD, GitHub Actions, deployment pipelines
   - Code refactoring, component restructuring
   - Database migrations, backend infrastructure
   - Internal API changes, gRPC, service communication
   - Developer tooling, linting, formatting
   - README updates, internal documentation
3. Write in simple, non-technical language
4. Focus on BENEFITS to the user, not implementation details
5. Group into these categories ONLY (skip empty categories):
   - **What's New** - New features users can now use
   - **Improved** - Enhancements to existing features
   - **Fixed** - Bugs that were affecting users
6. Start with a brief 1-2 sentence summary of the release highlights
7. Use bullet points, be concise (one line per item)
8. Do NOT wrap output in markdown code blocks
9. Do NOT include commit hashes or author names
10. If most commits are internal/technical, just summarize with 'Minor improvements and bug fixes'

Write the release notes directly in markdown format, ready to be used as-is."

# ─── Call ThreatWinds ─────────────────────────────────────────────────────────
echo "Calling ThreatWinds ($MODEL)..."
PAYLOAD=$(jq -n \
    --arg model "$MODEL" \
    --arg prompt "$PROMPT" \
    --argjson temp "$TEMPERATURE" \
    --argjson maxtok "$MAX_TOKENS" \
    '{
        model: $model,
        messages: [
            {role: "system", content: "You are a technical writer specializing in software changelogs."},
            {role: "user",   content: $prompt}
        ],
        temperature: $temp,
        max_tokens:  $maxtok
    }')

RESPONSE_FILE=$(mktemp)
HTTP_STATUS=$(curl -sS -o "$RESPONSE_FILE" -w '%{http_code}' \
    -X POST "${BASE_URL%/}/chat/completions" \
    -H "Content-Type: application/json" \
    -H "api-key: ${THREATWINDS_API_KEY}" \
    -H "api-secret: ${THREATWINDS_API_SECRET}" \
    --data "$PAYLOAD" || echo "000")

if [ "$HTTP_STATUS" != "200" ]; then
    echo "ERROR: ThreatWinds API returned HTTP $HTTP_STATUS" >&2
    cat "$RESPONSE_FILE" >&2
    exit 1
fi

CHANGELOG=$(jq -r '.choices[0].message.content // empty' "$RESPONSE_FILE")

if [ -z "$CHANGELOG" ]; then
    echo "ERROR: empty response from ThreatWinds." >&2
    cat "$RESPONSE_FILE" >&2
    exit 1
fi

# ─── Append comparison link ───────────────────────────────────────────────────
REPO_REMOTE=$(git config --get remote.origin.url 2>/dev/null | \
    sed -E 's#(git@github.com:|https://github.com/)#https://github.com/#; s#\.git$##')

CHANGELOG="${CHANGELOG}

---
**Full Changelog**: ${REPO_REMOTE}/compare/${PREVIOUS_TAG}...${CURRENT_TAG}"

printf "%s\n" "$CHANGELOG" > "$OUTPUT_FILE"

echo
echo "──────── GENERATED CHANGELOG ────────"
cat "$OUTPUT_FILE"
echo "─────────────────────────────────────"
echo
echo "Saved to: $OUTPUT_FILE"
