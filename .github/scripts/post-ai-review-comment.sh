#!/bin/bash
set -euo pipefail

# Reads every AI review result JSON (one per prompt, produced by
# ai-review.sh) plus the optional Go dependencies check output, and
# posts/updates a single sticky PR comment. Purely informational:
# severity/tier and the deps check result only drive the wording/icon of
# the comment. This never blocks the merge, never fails the job, and never
# @mentions anyone — it's just there so the author knows what to fix.
#
# Required env vars:
#   RESULTS_DIR         directory containing one <prompt-name>.json per
#                       AI prompt (as written by ai-review.sh)
#   PR_NUMBER           PR number to comment on
#   GITHUB_REPOSITORY   owner/repo
#   GITHUB_TOKEN        for posting/updating the comment (unless DRY_RUN=1)
#
# Optional env vars:
#   GO_DEPS_OUTPUT_FILE      path to go-deps.sh's captured stdout+stderr.
#                            If unset/missing, the Go dependencies section
#                            is omitted entirely.
#   GO_DEPS_EXIT_CODE_FILE   path to a file containing go-deps.sh's exit
#                            code. Required alongside GO_DEPS_OUTPUT_FILE.

: "${RESULTS_DIR:?RESULTS_DIR is required}"
: "${PR_NUMBER:?PR_NUMBER is required}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"

GO_DEPS_OUTPUT_FILE="${GO_DEPS_OUTPUT_FILE:-}"
GO_DEPS_EXIT_CODE_FILE="${GO_DEPS_EXIT_CODE_FILE:-}"

# DRY_RUN=1 prints the comment body instead of calling the GitHub API.
DRY_RUN="${DRY_RUN:-0}"

if [[ "$DRY_RUN" != "1" ]]; then
    : "${GITHUB_TOKEN:?GITHUB_TOKEN is required (or set DRY_RUN=1)}"
fi

# Same marker as the old approver.sh, so an already-open PR's existing
# comment gets updated in place instead of getting a duplicate.
MARKER='<!-- approver:ai -->'

api() {
    curl -sS \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$@"
}

find_sticky_comment() {
    api "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments?per_page=100" \
        | jq -r --arg m "$MARKER" '.[] | select(.body | contains($m)) | .id' \
        | head -n1
}

upsert_sticky_comment() {
    local body="$1"
    local full_body="${MARKER}"$'\n'"${body}"

    if [[ "$DRY_RUN" == "1" ]]; then
        echo "::group::[DRY_RUN] Would upsert AI review comment"
        echo "$full_body"
        echo "::endgroup::"
        return 0
    fi

    local id
    id=$(find_sticky_comment || true)
    if [[ -n "$id" ]]; then
        echo "Updating existing comment $id"
        jq -n --arg body "$full_body" '{body: $body}' \
            | api -X PATCH "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/comments/${id}" \
                --data-binary @- > /dev/null
    else
        echo "Creating new comment"
        jq -n --arg body "$full_body" '{body: $body}' \
            | api -X POST "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
                --data-binary @- > /dev/null
    fi
}

declare -a results=()
declare -i max_tier=1
has_block_sev=false      # any high/critical finding
has_any_findings=false   # any finding at all (for warning vs clean wording)
findings_md=""

shopt -s nullglob
for f in "$RESULTS_DIR"/*.json; do
    results+=("$f")

    tier=$(jq -r '.tier // 2' "$f")
    (( tier > max_tier )) && max_tier=$tier

    if jq -e '[(.findings // [])[].severity // "" | ascii_downcase] | any(. == "high" or . == "critical")' "$f" >/dev/null 2>&1; then
        has_block_sev=true
    fi
    if jq -e '((.findings // []) | length) > 0' "$f" >/dev/null 2>&1; then
        has_any_findings=true
    fi
done
shopt -u nullglob

no_results=false
if [[ ${#results[@]} -eq 0 ]]; then
    echo "::warning::No AI review result files found in $RESULTS_DIR"
    no_results=true
fi

# Build a markdown section per prompt result. Purely descriptive.
# Guarded on length first: "${results[@]}" on a zero-element array throws
# "unbound variable" under `set -u` on bash < 4.4 (e.g. macOS's default
# /bin/bash 3.2), even though the array itself is declared.
if (( ${#results[@]} > 0 )); then
    for f in "${results[@]}"; do
        prompt=$(jq -r '.prompt // "unknown"' "$f")
        model=$(jq -r '.model // "?"' "$f")
        summary=$(jq -r '.summary // "(no summary)"' "$f")
        p_block=$(jq -r '[(.findings // [])[].severity // "" | ascii_downcase] | any(. == "high" or . == "critical")' "$f" 2>/dev/null || echo false)
        p_count=$(jq -r '(.findings // []) | length' "$f" 2>/dev/null || echo 0)
        p_tier=$(jq -r '.tier // 2' "$f")
        findings=$(jq -r '
            (.findings // []) |
            if length == 0 then "  _No findings._"
            else
                map("  - **\(.severity // "?")** `\(.file // "?"):\(.line // "?")` — \(.message // "")") | join("\n")
            end
        ' "$f")
        if [[ "$p_block" == "true" || "$p_tier" == "3" ]]; then
            icon="🛑" label="high/critical — please review"
        elif (( p_count > 0 )); then
            icon="⚠️" label="minor findings"
        else
            icon="✅" label="clean"
        fi
        findings_md+=$'\n'"#### $icon \`$prompt\` (\`$model\`) — $label"$'\n\n'
        findings_md+="**Summary:** $summary"$'\n\n'
        findings_md+="$findings"$'\n'
    done
fi

# Go dependencies — same style as the AI prompt sections above, but not an
# AI call: just the exit code + output of go-deps.sh, ANSI colors stripped.
if [[ -n "$GO_DEPS_OUTPUT_FILE" && -f "$GO_DEPS_OUTPUT_FILE" && -n "$GO_DEPS_EXIT_CODE_FILE" && -f "$GO_DEPS_EXIT_CODE_FILE" ]]; then
    go_deps_exit=$(cat "$GO_DEPS_EXIT_CODE_FILE")
    go_deps_output=$(sed -E $'s/\x1b\\[[0-9;]*[mK]//g' "$GO_DEPS_OUTPUT_FILE")
    if [[ "$go_deps_exit" == "0" ]]; then
        findings_md+=$'\n'"#### 🟢 \`go-deps\` — up to date"$'\n\n'
        findings_md+="No pending Go dependency updates."$'\n'
    else
        findings_md+=$'\n'"#### 🔴 \`go-deps\` — pending updates"$'\n\n'
        findings_md+='```'$'\n'"$go_deps_output"$'\n''```'$'\n'
    fi
fi

if $no_results; then
    header="### ❓ AI review — could not run"
    intro="No AI results were produced. Check the workflow logs."
elif (( max_tier >= 3 )); then
    header="### 🛑 AI review — Sensitive area, extra care recommended"
    intro="This PR touches critical paths or introduces changes the model cannot judge with sufficient confidence. Review carefully before merging."
elif $has_block_sev; then
    header="### 🛑 AI review — High/critical findings"
    intro="One or more high/critical issues were found. Please review and fix before merging if they're real."
elif $has_any_findings; then
    header="### ✅ AI review — Minor findings"
    intro="Only minor (medium/low) issues were found. Consider addressing them."
else
    header="### ✅ AI review — Clean"
    intro="No issues detected in this diff."
fi

body=$(cat <<EOF
$header

$intro
$findings_md
EOF
)

upsert_sticky_comment "$body"

echo "Done. This comment is informational only — it never blocks the merge."
