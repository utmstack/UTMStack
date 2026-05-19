#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

print_usage() {
    cat << 'EOF'
Usage: go-deps.sh [--check|--update] [--discover|<path>]

Modes:
  --check     Check for outdated dependencies (exit 1 if found)
  --update    Update outdated dependencies (default)

Target:
  --discover  Discover all Go projects from current directory
  <path>      Path to a specific Go project

Examples:
  go-deps.sh --check ./installer
  go-deps.sh --update ./installer
  go-deps.sh --check --discover
  go-deps.sh --update --discover
EOF
}

# Parse arguments
CHECK_ONLY=false
DISCOVER=false
TARGET_PATH=""

if [[ $# -lt 1 ]]; then
    print_usage
    exit 1
fi

for arg in "$@"; do
    case "$arg" in
        --check)
            CHECK_ONLY=true
            ;;
        --update)
            CHECK_ONLY=false
            ;;
        --discover)
            DISCOVER=true
            ;;
        --help|-h)
            print_usage
            exit 0
            ;;
        --*)
            echo -e "${RED}Error: unknown option $arg${NC}" >&2
            print_usage
            exit 1
            ;;
        *)
            TARGET_PATH="$arg"
            ;;
    esac
done

# Validate arguments
if [[ "$DISCOVER" == false && -z "$TARGET_PATH" ]]; then
    echo -e "${RED}Error: must specify a path or use --discover${NC}" >&2
    print_usage
    exit 1
fi

if [[ "$DISCOVER" == true && -n "$TARGET_PATH" ]]; then
    echo -e "${RED}Error: cannot use both --discover and a specific path${NC}" >&2
    print_usage
    exit 1
fi

# Discover Go projects
discover_projects() {
    local root="$1"
    find "$root" -name "go.mod" \
        -not -path "*/.*" \
        -not -path "*/vendor/*" \
        -not -path "*/node_modules/*" \
        -exec dirname {} \;
}

# Get explicit modules from go.mod (direct dependencies only)
get_explicit_modules() {
    local go_mod="$1"
    grep -E '^\s+[a-zA-Z]' "$go_mod" 2>/dev/null | \
        grep -v '//' | \
        awk '{print $1}' | \
        sort -u
}

# Check for updates in a project, outputs JSON lines.
# On `go list` failure, records the project + error in $TEMP_DIR/_check_errors
# instead of returning non-zero, so the caller can keep checking the other
# projects and we report every broken module at once instead of one per run.
check_project() {
    local project_path="$1"
    local go_mod="$project_path/go.mod"

    # Get explicit modules
    local explicit_modules
    explicit_modules=$(get_explicit_modules "$go_mod")

    # Get all modules with update info as JSON
    local json_output
    local list_err
    list_err=$(mktemp)
    if ! json_output=$(cd "$project_path" && go list -u -m -json all 2>"$list_err"); then
        {
            echo "## $project_path"
            cat "$list_err"
            echo
        } >> "$TEMP_DIR/_check_errors"
        rm -f "$list_err"
        return 0
    fi
    rm -f "$list_err"

    # Parse JSON and filter modules with updates that are explicit dependencies
    echo "$json_output" | jq -c 'select(.Update != null) | {Path: .Path, Version: .Version, UpdateVersion: .Update.Version}' 2>/dev/null | \
    while IFS= read -r module; do
        local mod_path
        mod_path=$(echo "$module" | jq -r '.Path')
        if echo "$explicit_modules" | grep -qx "$mod_path"; then
            echo "$module"
        fi
    done
}

# Update a single module
update_module() {
    local project_path="$1"
    local mod_path="$2"
    local new_version="$3"

    local update_str="${mod_path}@${new_version}"
    echo -e "     🔄 Updating $update_str"
    (cd "$project_path" && go get "$update_str")
}

# Run go mod tidy
run_tidy() {
    local project_path="$1"
    echo -e "     🧹 Running go mod tidy..."
    (cd "$project_path" && go mod tidy)
}

# Create temp directory for storing updates per project
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Get list of projects
PROJECTS=""
if [[ "$DISCOVER" == true ]]; then
    PROJECTS=$(discover_projects ".")

    if [[ -z "$PROJECTS" ]]; then
        echo "No Go projects found."
        exit 0
    fi

    project_count=$(echo "$PROJECTS" | wc -l | tr -d ' ')
    echo -e "🔍 Discovered $project_count Go projects\n"
else
    if [[ ! -f "$TARGET_PATH/go.mod" ]]; then
        echo -e "${RED}Error: no go.mod found in $TARGET_PATH${NC}" >&2
        exit 1
    fi
    PROJECTS="$TARGET_PATH"
fi

# Check all projects for updates
HAS_UPDATES=false

echo "$PROJECTS" | while IFS= read -r project; do
    [[ -z "$project" ]] && continue

    updates=$(check_project "$project")
    if [[ -n "$updates" ]]; then
        # Store updates in a temp file named after the project (sanitized)
        safe_name=$(echo "$project" | tr '/' '_')
        echo "$updates" > "$TEMP_DIR/$safe_name"
        echo "$project" >> "$TEMP_DIR/_projects_with_updates"
    fi
done

# If any project couldn't be inspected, report all of them at once.
# This is almost always caused by stale go.sum entries in modules that
# consume an internal package via `replace` — fix those first before
# anything else, since both --check and --update need a clean read.
if [[ -f "$TEMP_DIR/_check_errors" ]]; then
    echo -e "${RED}❌ Could not inspect the following projects (run 'go mod tidy' there):${NC}" >&2
    echo >&2
    cat "$TEMP_DIR/_check_errors" >&2
    exit 1
fi

# Check if any updates were found
if [[ ! -f "$TEMP_DIR/_projects_with_updates" ]]; then
    echo -e "${GREEN}✅ All dependencies are up to date.${NC}"
    exit 0
fi

# Print summary of updates needed
echo -e "📦 Dependencies with updates available:"
while IFS= read -r project; do
    [[ -z "$project" ]] && continue
    safe_name=$(echo "$project" | tr '/' '_')

    echo -e "\n  📁 $project:"
    while IFS= read -r module; do
        [[ -z "$module" ]] && continue
        mod_path=$(echo "$module" | jq -r '.Path')
        current=$(echo "$module" | jq -r '.Version')
        new_ver=$(echo "$module" | jq -r '.UpdateVersion')
        echo -e "     - $mod_path: $current → $new_ver"
    done < "$TEMP_DIR/$safe_name"
done < "$TEMP_DIR/_projects_with_updates"

if [[ "$CHECK_ONLY" == true ]]; then
    echo -e "\n${RED}❌ Please update dependencies before merging.${NC}"
    exit 1
fi

# Update mode - apply updates
echo -e "\n🔄 Updating dependencies..."
while IFS= read -r project; do
    [[ -z "$project" ]] && continue
    safe_name=$(echo "$project" | tr '/' '_')

    echo -e "\n  📁 $project:"
    while IFS= read -r module; do
        [[ -z "$module" ]] && continue
        mod_path=$(echo "$module" | jq -r '.Path')
        new_ver=$(echo "$module" | jq -r '.UpdateVersion')
        update_module "$project" "$mod_path" "$new_ver"
    done < "$TEMP_DIR/$safe_name"

    run_tidy "$project"
done < "$TEMP_DIR/_projects_with_updates"

echo -e "\n${GREEN}✅ All dependencies updated successfully.${NC}"

# Propagate `go mod tidy` to EVERY discovered project, not just those that
# were updated. When an internal package (e.g. packages/go-common) gets new
# transitive deps, every consumer that imports it via `replace ../packages/...`
# needs its own go.sum recomputed — otherwise the next `go list` call fails.
echo -e "\n🧹 Propagating tidy to all projects (handles local replace ripple)..."
echo "$PROJECTS" | while IFS= read -r project; do
    [[ -z "$project" ]] && continue
    echo -e "  📁 $project"
    (cd "$project" && go mod tidy)
done

# Verify every project still builds. We compile into a throwaway directory so
# main packages don't litter the working tree, then delete it.
echo -e "\n🔨 Verifying all projects build..."
BUILD_OUT=$(mktemp -d)
BUILD_FAILED_FILE="$TEMP_DIR/_build_failed"
echo "$PROJECTS" | while IFS= read -r project; do
    [[ -z "$project" ]] && continue
    if (cd "$project" && go build -o "$BUILD_OUT/" ./... 2>&1); then
        echo -e "  ${GREEN}✅${NC} $project"
    else
        echo -e "  ${RED}❌${NC} $project"
        echo "$project" >> "$BUILD_FAILED_FILE"
    fi
done
rm -rf "$BUILD_OUT"

if [[ -f "$BUILD_FAILED_FILE" ]]; then
    echo -e "\n${RED}❌ Build failed in:${NC}" >&2
    cat "$BUILD_FAILED_FILE" >&2
    exit 1
fi

echo -e "\n${GREEN}✅ All projects build cleanly.${NC}"
