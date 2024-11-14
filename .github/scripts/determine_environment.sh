#!/bin/bash

if [ "$GITHUB_EVENT_NAME" == "push" ] && [[ "$GITHUB_REF" == refs/heads/feature/* ]]; then
    echo "DEV environment"
    echo "env_tag=dev" >> "$GITHUB_OUTPUT"
elif [ "$GITHUB_EVENT_NAME" == "pull_request_review" ] && [ "$GITHUB_REVIEW_STATE" == "approved" ] && [ "$GITHUB_BASE_REF" == "main" ] && [[ "$GITHUB_HEAD_REF" == feature/* ]]; then
    echo "QA environment"
    echo "env_tag=qa" >> "$GITHUB_OUTPUT"
elif [ "$GITHUB_EVENT_NAME" == "push" ] && [ "$GITHUB_REF" == refs/heads/main ]; then
    echo "RC environment"
    echo "env_tag=rc" >> "$GITHUB_OUTPUT"
elif [ "$GITHUB_EVENT_NAME" == "release" ]; then
    echo "PRODUCTION environment"
    echo "env_tag=prod" >> "$GITHUB_OUTPUT"
fi
