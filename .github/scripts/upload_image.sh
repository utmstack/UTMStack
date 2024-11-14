#!/bin/bash

if [ -z "$1" ]; then
  echo "No services provided to send."
  exit 1
fi

image_services=$(echo "$1" | jq -r '.[]')
echo "Sending the following image services to the server: $image_services"

workspace="$GITHUB_WORKSPACE"
url="$CM_API"

echo "Workspace: $workspace"
echo "URL: $url"

auth_id=$(echo "$CM_PUB_AUTH" | jq -r '.id')
auth_key=$(echo "$CM_PUB_AUTH" | jq -r '.key')

versions_json_path="$workspace/versions.json"
if [ ! -f "$versions_json_path" ]; then
    echo "Error: versions.json not found at $versions_json_path"
    exit 1
fi

versions_content=$(cat "$versions_json_path")
echo "Versions: $versions_content"

master_changelog_path="$workspace/CHANGELOG.md"
if [ ! -f "$master_changelog_path" ]; then
    echo "Error: CHANGELOG.md not found at $master_changelog_path"
    exit 1
fi
master_changelog=$(cat "$master_changelog_path")

master_version_data=$(jq -n \
    --arg changelog "$master_changelog" \
    --arg version_name "$(echo "$versions_content" | jq -r '.version')" \
    '{changelog: $changelog, version_name: $version_name}')

echo "Sending main version..."
master_version_response=$(curl -s -X POST "$url/master-version" \
    -H "publisher-key: $auth_key" \
    -H "publisher-id: $auth_id" \
    -H "Content-Type: application/json" \
    -d "$master_version_data")

master_version_id="$master_version_response"
echo "Master Version ID: $master_version_id"

for service in $image_services; do
    echo "Processing service: $service"
    version=$(echo "$versions_content" | jq -r --arg s "$service" '.[$s]')
    service_path="$workspace/$service"

    changelog_path="$service_path/CHANGELOG.md"
    readme_path="$service_path/README.md"

    if [ ! -f "$changelog_path" ] || [ ! -f "$readme_path" ]; then
        echo "Error: Missing CHANGELOG.md or README.md for service $name"
        echo "CHANGELOG.md: $changelog_path"
        echo "README.md: $readme_path"
        exit 1
    fi

    changelog=$(cat "$changelog_path")
    readme=$(cat "$readme_path")

    component_version_data=$(jq -n \
        --arg changelog "$changelog" \
        --arg description "$readme" \
        --arg master_version_id "$master_version_id" \
        --arg name "$service" \
        --arg version_name "$version" \
        '{changelog: $changelog, description: $description, editions: ["Community", "Enterprise"], master_version_id: $master_version_id, name: $name, version_name: $version_name}')

    component_version_response=$(curl -s -X POST "$url/component-version" \
        -H "publisher-key: $auth_key" \
        -H "publisher-id: $auth_id" \
        -H "Content-Type: application/json" \
        -d "$component_version_data")

    component_version_id="$component_version_response"
    echo "Component Version ID: $component_version_id"
done
