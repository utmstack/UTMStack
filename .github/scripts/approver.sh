#!/bin/bash
set -euo pipefail

# Consolidates artifacts from go_deps and ai_review and decides whether the
# PR passes. Posts sticky comments and (if APPROVER_TOKEN is provided) leaves
# a formal PR review.
#
# Required env vars:
#   ARTIFACTS_DIR         directory where workflow downloaded all artifacts
#   PR_NUMBER             PR number to comment on
#   GITHUB_REPOSITORY     owner/repo
#   GITHUB_TOKEN          for posting/updating comments (always)
#
# Optional env vars:
#   APPROVER_TOKEN        PAT or app token with `pull_requests: write` and the
#                         ability to approve reviews. If missing, the approver
#                         only posts comments and sets the check status.
#   TIER3_REVIEWERS       comma-separated GitHub handles to @mention on tier 3
#                         (without @ prefix; e.g. "alice,bob")
#   API_SECRET            PAT with `read:org` to check team membership for
#                         administrators and core-developers. If missing, the
#                         permission check is skipped (treated as authorized).
#   PR_AUTHOR             GitHub login of the PR author. Required for the
#                         permission check.
#   BASE_REF              PR target branch (e.g. "release/v11.2.9"). Required;
#                         auto-merge only fires when this starts with "release/".
#   ORG                   GitHub org to look up teams in. Default: "utmstack".
#   ADMIN_TEAM            Team slug for administrators. Default: "administrators".
#   CORE_TEAM             Team slug for core-developers. Default: "core-developers".
#   MERGE_METHOD          One of "merge"|"squash"|"rebase". Default: "squash".

: "${ARTIFACTS_DIR:?ARTIFACTS_DIR is required}"
: "${PR_NUMBER:?PR_NUMBER is required}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"

# DRY_RUN=1 makes the approver print the bodies it would post and skip all
# GitHub API calls (no comments, no review). Useful for local testing.
DRY_RUN="${DRY_RUN:-0}"

if [[ "$DRY_RUN" != "1" ]]; then
    : "${GITHUB_TOKEN:?GITHUB_TOKEN is required (or set DRY_RUN=1)}"
fi

APPROVER_TOKEN="${APPROVER_TOKEN:-}"
TIER3_REVIEWERS="${TIER3_REVIEWERS:-}"
API_SECRET="${API_SECRET:-}"
PR_AUTHOR="${PR_AUTHOR:-}"
BASE_REF="${BASE_REF:-}"
ORG="${ORG:-utmstack}"
ADMIN_TEAM="${ADMIN_TEAM:-administrators}"
CORE_TEAM="${CORE_TEAM:-core-developers}"
MERGE_METHOD="${MERGE_METHOD:-squash}"

# Markers for sticky comments. Each topic has its own marker so deps, AI and
# permission updates don't trample each other.
MARKER_DEPS='<!-- approver:deps -->'
MARKER_AI='<!-- approver:ai -->'
MARKER_PERM='<!-- approver:permission -->'

api() {
    curl -sS \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$@"
}

# Find sticky comment ID by marker substring, or empty if none.
find_sticky_comment() {
    local marker="$1"
    api "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments?per_page=100" \
        | jq -r --arg m "$marker" '.[] | select(.body | contains($m)) | .id' \
        | head -n1
}

# Upsert a sticky comment: edit if marker already present, else create.
upsert_sticky_comment() {
    local marker="$1"
    local body="$2"
    local full_body="${marker}"$'\n'"${body}"

    if [[ "$DRY_RUN" == "1" ]]; then
        echo "::group::[DRY_RUN] Would upsert comment with marker $marker"
        echo "$full_body"
        echo "::endgroup::"
        return 0
    fi

    local id
    id=$(find_sticky_comment "$marker" || true)

    if [[ -n "$id" ]]; then
        echo "Updating existing comment $id"
        jq -n --arg body "$full_body" '{body: $body}' \
            | api -X PATCH \
                "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/comments/${id}" \
                --data-binary @- > /dev/null
    else
        echo "Creating new comment"
        jq -n --arg body "$full_body" '{body: $body}' \
            | api -X POST \
                "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
                --data-binary @- > /dev/null
    fi
}

# Delete a sticky comment if it exists (used when the topic becomes a no-op).
delete_sticky_comment() {
    local marker="$1"
    if [[ "$DRY_RUN" == "1" ]]; then
        echo "[DRY_RUN] Would delete stale comment with marker $marker (if present)"
        return 0
    fi
    local id
    id=$(find_sticky_comment "$marker" || true)
    if [[ -n "$id" ]]; then
        echo "Deleting stale comment $id"
        api -X DELETE \
            "https://api.github.com/repos/${GITHUB_REPOSITORY}/issues/comments/${id}" \
            > /dev/null
    fi
}

# =============================================================================
# 1. Deps verdict
# =============================================================================

deps_artifact_dir="$ARTIFACTS_DIR/go-deps-result"
deps_failed=false
deps_output=""

if [[ -f "$deps_artifact_dir/exit_code.txt" ]]; then
    deps_exit=$(cat "$deps_artifact_dir/exit_code.txt")
    deps_output=$(cat "$deps_artifact_dir/output.txt" 2>/dev/null || echo "")
    if [[ "$deps_exit" != "0" ]]; then
        deps_failed=true
    fi
else
    echo "::warning::go-deps artifact missing — treating as failed"
    deps_failed=true
    deps_output="(go-deps artifact missing — the job may have failed to run)"
fi

# =============================================================================
# 2. AI verdict — read every ai-review-* artifact
# =============================================================================

declare -a ai_results=()
declare -i max_tier=1
ai_findings_md=""

shopt -s nullglob
for d in "$ARTIFACTS_DIR"/ai-review-*/; do
    f="${d}result.json"
    [[ -f "$f" ]] || continue
    ai_results+=("$f")
    tier=$(jq -r '.tier // 2' "$f")
    if (( tier > max_tier )); then
        max_tier=$tier
    fi
done
shopt -u nullglob

if [[ ${#ai_results[@]} -eq 0 ]]; then
    echo "::warning::No AI review artifacts — treating as tier 2 fail-safe"
    max_tier=2
fi

# Build a markdown section per AI prompt result.
for f in "${ai_results[@]}"; do
    prompt=$(jq -r '.prompt // "unknown"' "$f")
    model=$(jq -r '.model // "?"' "$f")
    tier=$(jq -r '.tier // 2' "$f")
    summary=$(jq -r '.summary // "(no summary)"' "$f")
    findings=$(jq -r '
        .findings // [] |
        if length == 0 then "  _No findings._"
        else
            map("  - **\(.severity // "?")** `\(.file // "?"):\(.line // "?")` — \(.message // "")") | join("\n")
        end
    ' "$f")
    case "$tier" in
        1) icon="✅" label="Tier 1 — looks clean" ;;
        2) icon="⚠️" label="Tier 2 — changes requested" ;;
        3) icon="🛑" label="Tier 3 — engineer review required" ;;
        *) icon="❓" label="Tier ?" ;;
    esac
    ai_findings_md+=$'\n'"#### $icon \`$prompt\` (\`$model\`) — $label"$'\n\n'
    ai_findings_md+="**Summary:** $summary"$'\n\n'
    ai_findings_md+="$findings"$'\n'
done

# =============================================================================
# 3. Compose & post comments
# =============================================================================

# --- Deps comment (only when failed) ---
if $deps_failed; then
    deps_body=$(cat <<EOF
### ❌ Go dependencies check failed

There are outdated Go dependencies, or modules that could not be inspected.
Run \`bash .github/scripts/go-deps.sh --update --discover\` locally and
commit the updated \`go.mod\` / \`go.sum\` files.

<details><summary>Script output</summary>

\`\`\`
$deps_output
\`\`\`

</details>
EOF
)
    upsert_sticky_comment "$MARKER_DEPS" "$deps_body"
else
    # Remove any previous deps comment now that deps are clean.
    delete_sticky_comment "$MARKER_DEPS"
fi

# --- AI verdict comment (always) ---
case "$max_tier" in
    1)
        ai_header="### ✅ AI review — Approved"
        ai_intro="All prompts returned Tier 1. No blocking issues detected in this diff."
        ;;
    2)
        ai_header="### ⚠️ AI review — Changes requested"
        ai_intro="One or more prompts found issues the author should fix before merging. Details below."
        ;;
    3)
        mention=""
        if [[ -n "$TIER3_REVIEWERS" ]]; then
            IFS=',' read -ra handles <<< "$TIER3_REVIEWERS"
            for h in "${handles[@]}"; do
                h="$(echo "$h" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e 's/^@//')"
                [[ -n "$h" ]] && mention+="@$h "
            done
        fi
        ai_header="### 🛑 AI review — Engineer review required"
        ai_intro="This PR touches critical paths or introduces changes the model cannot judge with sufficient confidence. ${mention}please review."
        ;;
    *)
        ai_header="### ❓ AI review — Unknown verdict"
        ai_intro="The approver could not determine an overall tier."
        ;;
esac

ai_body=$(cat <<EOF
$ai_header

$ai_intro
$ai_findings_md
EOF
)

upsert_sticky_comment "$MARKER_AI" "$ai_body"

# =============================================================================
# 4. Permission check — LAST gate before approval.
# We always post the deps + AI comments above (regardless of who opened the
# PR), so unauthorized contributors still see what would need to be fixed.
# But authorization is required to actually get the approval / auto-merge.
# =============================================================================

is_in_team() {
    local team="$1"
    local user="$2"
    [[ -z "$API_SECRET" ]] && return 1
    [[ -z "$user" ]] && return 1
    local resp
    resp=$(curl -sS \
        -H "Authorization: Bearer ${API_SECRET}" \
        -H "Accept: application/vnd.github+json" \
        "https://api.github.com/orgs/${ORG}/teams/${team}/memberships/${user}")
    echo "$resp" | jq -e '.state == "active"' >/dev/null 2>&1
}

authorized=false
if [[ -z "$API_SECRET" ]]; then
    echo "::warning::API_SECRET not set — skipping permission check (treating as authorized)"
    authorized=true
elif [[ -z "$PR_AUTHOR" ]]; then
    echo "::warning::PR_AUTHOR not set — skipping permission check (treating as authorized)"
    authorized=true
elif is_in_team "$ADMIN_TEAM" "$PR_AUTHOR"; then
    echo "✅ $PR_AUTHOR is in @${ORG}/${ADMIN_TEAM}"
    authorized=true
elif is_in_team "$CORE_TEAM" "$PR_AUTHOR"; then
    echo "✅ $PR_AUTHOR is in @${ORG}/${CORE_TEAM}"
    authorized=true
else
    echo "⛔ $PR_AUTHOR is NOT in @${ORG}/${ADMIN_TEAM} nor @${ORG}/${CORE_TEAM}"
fi

if ! $authorized; then
    perm_body=$(cat <<EOF
### ⛔ Permission denied

Only members of @${ORG}/${ADMIN_TEAM} or @${ORG}/${CORE_TEAM} can merge PRs into this repository.

**PR author:** @${PR_AUTHOR}

The comments above (deps + AI review) are still valid — if you address them, the checks will re-run automatically, but final approval is left to an administrator.

@${ORG}/${ADMIN_TEAM} please review.
EOF
)
    upsert_sticky_comment "$MARKER_PERM" "$perm_body"
else
    # Clear the permission comment if it existed from a previous run.
    delete_sticky_comment "$MARKER_PERM"
fi

# =============================================================================
# 5. Formal PR review (only when APPROVER_TOKEN is present)
# =============================================================================

if [[ -n "$APPROVER_TOKEN" || "$DRY_RUN" == "1" ]]; then
    if ! $deps_failed && [[ "$max_tier" == "1" ]] && $authorized; then
        review_event="APPROVE"
        review_body="Approved by AI review (Tier 1, deps OK, authorized author)."
    elif ! $authorized; then
        review_event="REQUEST_CHANGES"
        review_body="Author is not in @${ORG}/${ADMIN_TEAM} nor @${ORG}/${CORE_TEAM} — admin review required."
    elif [[ "$max_tier" == "3" ]] || $deps_failed; then
        review_event="REQUEST_CHANGES"
        review_body="Changes requested — see approver comments above."
    else
        review_event="REQUEST_CHANGES"
        review_body="Changes requested by AI review (Tier 2)."
    fi

    if [[ "$DRY_RUN" == "1" ]]; then
        echo "[DRY_RUN] Would post formal PR review: event=$review_event body=\"$review_body\""
    else
        echo "Posting formal PR review: $review_event"
        payload=$(jq -n \
            --arg event "$review_event" \
            --arg body "$review_body" \
            '{event: $event, body: $body}')
        curl -sS \
            -H "Authorization: Bearer ${APPROVER_TOKEN}" \
            -H "Accept: application/vnd.github+json" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            -X POST \
            "https://api.github.com/repos/${GITHUB_REPOSITORY}/pulls/${PR_NUMBER}/reviews" \
            --data "$payload" > /dev/null || echo "::warning::Failed to post formal review"
    fi
else
    echo "::warning::APPROVER_TOKEN not set — skipping formal PR review (only sticky comments + status check)"
fi

# =============================================================================
# 6. Auto-merge — only when target branch is release/**.
# We enable GitHub native auto-merge (gh pr merge --auto), which queues the
# merge until ALL branch-protection requirements are met. It's safe: if any
# other check fails or someone leaves a manual REQUEST_CHANGES, the merge
# stays pending.
# =============================================================================

if ! $deps_failed && [[ "$max_tier" == "1" ]] && $authorized; then
    if [[ "$BASE_REF" == release/* ]]; then
        if [[ "$DRY_RUN" == "1" ]]; then
            echo "[DRY_RUN] Would enable auto-merge: gh pr merge $PR_NUMBER --auto --$MERGE_METHOD (base: $BASE_REF)"
        elif [[ -z "$APPROVER_TOKEN" ]]; then
            echo "::warning::APPROVER_TOKEN not set — cannot enable auto-merge"
        else
            echo "Enabling auto-merge for #$PR_NUMBER (target: $BASE_REF, method: $MERGE_METHOD)"
            GH_TOKEN="$APPROVER_TOKEN" gh pr merge "$PR_NUMBER" \
                --auto "--${MERGE_METHOD}" \
                --repo "$GITHUB_REPOSITORY" \
                || echo "::warning::Failed to enable auto-merge (already enabled? branch protection mismatch?)"
        fi
    else
        echo "Target branch '$BASE_REF' is not release/** — skipping auto-merge (deploy branches stay manual)"
    fi
fi

# =============================================================================
# 7. Exit code
# =============================================================================

echo ""
echo "Summary:"
echo "  deps_failed: $deps_failed"
echo "  max_tier:    $max_tier"
echo "  authorized:  $authorized"
echo "  base_ref:    $BASE_REF"

if $deps_failed; then
    exit 1
fi

if ! $authorized; then
    exit 1
fi

if [[ "$max_tier" -ge 2 ]]; then
    exit 1
fi

echo "✅ PR approved."
exit 0
